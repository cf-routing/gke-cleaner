package poller

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/christianang/gke-cleaner/pkg/store"
	"github.com/go-logr/logr"

	container "cloud.google.com/go/container/apiv1"
	containerpb "google.golang.org/genproto/googleapis/container/v1"
)

type GKE struct {
	Log          logr.Logger
	Client       *container.ClusterManagerClient
	ClusterStore *store.Cluster

	Project                string
	PollInterval           time.Duration
	LifetimeDuration       time.Duration
	ResourceLabelFilterMap []string
}

func (g *GKE) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	close(ready)
	ctx, cancel := context.WithCancel(context.Background())
	for {
		select {
		case <-signals:
			cancel()
			return nil
		case <-time.After(g.PollInterval):
			g.Log.V(1).Info("Polling")
			if err := g.syncGKEClusters(ctx); err != nil {
				g.Log.Error(err, "Failed to sync gke clusters")
				continue
			}

			if err := g.cleanupExpiredClusters(ctx); err != nil {
				g.Log.Error(err, "Failed to cleanup expired clusters")
				continue
			}
		}
	}
}

func (g *GKE) cleanupExpiredClusters(ctx context.Context) error {
	expiredClusters, err := g.ClusterStore.ListExpired(ctx)
	if err != nil {
		return err
	}

	listClustersResponse, err := g.Client.ListClusters(ctx, &containerpb.ListClustersRequest{
		Parent: fmt.Sprintf("projects/%s/locations/-", g.Project),
	})
	if err != nil {
		return err
	}

	for _, cluster := range expiredClusters {
		if cluster.Ignore {
			continue
		}

		location, err := g.getClusterLocation(listClustersResponse, cluster.Name)
		if err != nil {
			g.Log.Error(err, "Failed to get cluster location. Skipping.", "cluster", cluster.Name)
			continue
		}

		_, err = g.Client.DeleteCluster(ctx, &containerpb.DeleteClusterRequest{
			Name: fmt.Sprintf("projects/%s/locations/%s/clusters/%s", g.Project, location, cluster.Name),
		})
		if err != nil {
			g.Log.Error(err, "Failed to delete cluster. Skipping.", "cluster", cluster.Name)
			continue
		}
		g.Log.Info("Removed expired cluster", "cluster", cluster.Name)
	}

	return nil
}

func (g *GKE) getClusterLocation(response *containerpb.ListClustersResponse, name string) (string, error) {
	for _, cluster := range response.Clusters {
		if name == cluster.Name {
			return cluster.Location, nil
		}
	}

	return "", fmt.Errorf("couldn't find cluster: %s", name)
}

func (g *GKE) syncGKEClusters(ctx context.Context) error {
	response, err := g.Client.ListClusters(ctx, &containerpb.ListClustersRequest{
		Parent: fmt.Sprintf("projects/%s/locations/-", g.Project),
	})
	if err != nil {
		return err
	}

	knownClusters, err := g.ClusterStore.List(ctx)
	if err != nil {
		return err
	}

	clusters := filter(response.Clusters, g.ResourceLabelFilterMap)

	addedClusters, removedClusters, updatedClusters := diffClusters(clusters, knownClusters)

	for _, cluster := range addedClusters {
		g.Log.Info("Discovered", "cluster", cluster)
		createTime, err := time.Parse(time.RFC3339, cluster.GetCreateTime())
		if err != nil {
			return err
		}

		err = g.ClusterStore.Insert(ctx, cluster.GetName(), createTime, createTime.Add(g.LifetimeDuration), false)
		if err != nil {
			return err
		}
	}

	for _, cluster := range updatedClusters {
		g.Log.Info("Updated", "cluster", cluster)
		createTime, err := time.Parse(time.RFC3339, cluster.GetCreateTime())
		if err != nil {
			return err
		}

		g.Log.Info("Update cluster", "clusterName", cluster.GetName(), "createTime", createTime, "expirationDate", createTime.Add(g.LifetimeDuration))
		err = g.ClusterStore.UpdateCreateAndExpirationDate(ctx, cluster.GetName(), createTime, createTime.Add(g.LifetimeDuration))
		if err != nil {
			return err
		}
	}

	for _, cluster := range removedClusters {
		g.Log.Info("Detected removal", "cluster", cluster)
		err = g.ClusterStore.Delete(ctx, cluster.GetName())
		if err != nil {
			return err
		}
	}

	return nil
}

func filter(clusters []*containerpb.Cluster, filters []string) []*containerpb.Cluster {
	filteredClusters := []*containerpb.Cluster{}
	for _, cluster := range clusters {
		for _, filter := range filters {
			k, v := parseFilter(filter)
			if cluster.ResourceLabels[k] == v {
				filteredClusters = append(filteredClusters, cluster)
				break
			}
		}
	}

	return filteredClusters
}

func parseFilter(filter string) (string, string) {
	s := strings.Split(filter, "=")
	return s[0], s[1]
}

func diffClusters(gkeClusters []*containerpb.Cluster, knownClusters []store.ClusterRecord) ([]Cluster, []Cluster, []Cluster) {
	added := []Cluster{}
	removed := []Cluster{}
	updated := []Cluster{}

	gkeClusterMap := map[string]*containerpb.Cluster{}
	for _, cluster := range gkeClusters {
		gkeClusterMap[cluster.Name] = cluster
	}

	knownClusterMap := map[string]*store.ClusterRecord{}
	for _, cluster := range knownClusters {
		knownClusterMap[cluster.Name] = &cluster
	}

	for clusterName, cluster := range gkeClusterMap {
		if _, found := knownClusterMap[clusterName]; !found {
			added = append(added, cluster)
		}
	}

	for clusterName, cluster := range knownClusterMap {
		c, found := gkeClusterMap[clusterName]
		if !found {
			removed = append(removed, c)
		} else if cluster.GetCreateTime() != c.GetCreateTime() {
			updated = append(updated, c)
		}
	}

	return added, removed, updated
}

type Cluster interface {
	GetName() string
	GetCreateTime() string
}
