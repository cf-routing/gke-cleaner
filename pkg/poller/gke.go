package poller

import (
	"context"
	"fmt"
	"os"
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

	Project          string
	PollInterval     time.Duration
	LifetimeDuration time.Duration
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

	for _, cluster := range expiredClusters {
		g.Log.Info("Removed expired cluster", "cluster", cluster.Name)
	}

	return nil
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

	addedClusters, removedClusters := diffClusters(response.Clusters, knownClusters)

	for _, cluster := range addedClusters {
		g.Log.Info("Discovered", "cluster", cluster)
		err = g.ClusterStore.Insert(ctx, cluster, time.Now().Add(g.LifetimeDuration), false)
		if err != nil {
			return err
		}
	}

	for _, cluster := range removedClusters {
		g.Log.Info("Detected removal", "cluster", cluster)
		err = g.ClusterStore.Delete(ctx, cluster)
		if err != nil {
			return err
		}
	}

	return nil
}

func diffClusters(gkeClusters []*containerpb.Cluster, knownClusters []store.ClusterRecord) ([]string, []string) {
	added := []string{}
	removed := []string{}

	gkeClusterMap := map[string]string{}
	for _, cluster := range gkeClusters {
		gkeClusterMap[cluster.Name] = ""
	}

	knownClusterMap := map[string]string{}
	for _, cluster := range knownClusters {
		knownClusterMap[cluster.Name] = ""
	}

	for cluster := range gkeClusterMap {
		if _, found := knownClusterMap[cluster]; !found {
			added = append(added, cluster)
		}
	}

	for cluster := range knownClusterMap {
		if _, found := gkeClusterMap[cluster]; !found {
			removed = append(removed, cluster)
		}
	}

	return added, removed
}
