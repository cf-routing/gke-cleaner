package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/christianang/gke-cleaner/pkg/store"
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
)

type Cluster struct {
	Log              logr.Logger
	ClusterStore     *store.Cluster
	LifetimeDuration time.Duration
}

func (c *Cluster) List(w http.ResponseWriter, req *http.Request) {
	knownClusters, err := c.ClusterStore.List(context.Background())
	if err != nil {
		c.Log.Error(err, "failed to list clusters")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body, err := json.Marshal(knownClusters)
	if err != nil {
		c.Log.Error(err, "failed to marshal clusters")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(body)
	if err != nil {
		c.Log.Error(err, "failed to write to response body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (c *Cluster) Renew(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	err := c.ClusterStore.UpdateExpirationDate(context.Background(), vars["name"], time.Now().Add(c.LifetimeDuration))
	if err != nil {
		c.Log.Error(err, "failed to update ignore")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (c *Cluster) Ignore(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	err := c.ClusterStore.UpdateIgnore(context.Background(), vars["name"], true)
	if err != nil {
		c.Log.Error(err, "failed to update ignore")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (c *Cluster) Unignore(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	err := c.ClusterStore.UpdateIgnore(context.Background(), vars["name"], false)
	if err != nil {
		c.Log.Error(err, "failed to update ignore")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
