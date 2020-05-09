package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"syscall"

	"github.com/christianang/gke-cleaner/pkg/config"
	"github.com/christianang/gke-cleaner/pkg/handler"
	"github.com/christianang/gke-cleaner/pkg/migrate"
	"github.com/christianang/gke-cleaner/pkg/poller"
	"github.com/christianang/gke-cleaner/pkg/store"
	"github.com/go-logr/zapr"
	"github.com/gorilla/mux"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/sigmon"
	"go.uber.org/zap"

	container "cloud.google.com/go/container/apiv1"
	ifrithttpserver "github.com/tedsuo/ifrit/http_server"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	zapLog, err := zap.NewDevelopment()
	if err != nil {
		fmt.Printf("failed to setup logger: %s\n", err)
		os.Exit(1)
	}

	log := zapr.NewLogger(zapLog)

	cfg, err := config.LoadFromEnv(log.WithName("config.LoadFromEnv"))
	if err != nil {
		log.WithName("main").Error(err, "failed to load config")
		os.Exit(1)
	}

	db, err := sql.Open("sqlite3", "./temp.db")
	if err != nil {
		log.WithName("main").Error(err, "failed to open connection to database")
		os.Exit(1)
	}

	migrateDB := &migrate.DB{
		Log: log.WithName("migrate.DB"),
		DB:  db,
	}

	clusterStore := &store.Cluster{
		DB: db,
	}

	clusterHandler := &handler.Cluster{
		Log:              log.WithName("handler.Cluster"),
		ClusterStore:     clusterStore,
		LifetimeDuration: cfg.ClusterLifetimeDuration,
	}

	router := mux.NewRouter()
	router.HandleFunc("/clusters", clusterHandler.List)
	router.HandleFunc("/clusters/renew/{name}", clusterHandler.Renew).Methods("POST")
	router.HandleFunc("/clusters/ignore/{name}", clusterHandler.Ignore).Methods("POST")
	router.HandleFunc("/clusters/unignore/{name}", clusterHandler.Unignore).Methods("POST")

	server := ifrithttpserver.New(fmt.Sprintf(":%d", cfg.Port), router)

	clusterManagerClient, err := container.NewClusterManagerClient(context.Background())
	if err != nil {
		log.WithName("main").Error(err, "failed to create gcloud cluster manager client")
		os.Exit(1)
	}

	gkePoller := &poller.GKE{
		Log:              log.WithName("poller.GKE"),
		Client:           clusterManagerClient,
		ClusterStore:     clusterStore,
		Project:          cfg.Project,
		PollInterval:     cfg.GCloudPollInterval,
		LifetimeDuration: cfg.ClusterLifetimeDuration,
	}

	group := grouper.NewOrdered(os.Interrupt, grouper.Members{
		grouper.Member{Name: "migrate-db", Runner: migrateDB},
		grouper.Member{Name: "server", Runner: server},
		grouper.Member{Name: "gke-poller", Runner: gkePoller},
	})

	monitor := ifrit.Invoke(sigmon.New(group, syscall.SIGTERM, syscall.SIGINT))

	log.WithName("main").Info("Started")
	err = <-monitor.Wait()
	if err != nil {
		log.WithName("main").Error(err, "failed to start all processes")
		os.Exit(1)
	}
}
