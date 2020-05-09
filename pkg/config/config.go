package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/go-logr/logr"
)

type Config struct {
	Port                    int
	Project                 string
	GCloudPollInterval      time.Duration
	ClusterLifetimeDuration time.Duration
}

func LoadFromEnv(log logr.Logger) (Config, error) {
	portStr, ok := os.LookupEnv("PORT")
	if !ok {
		return Config{}, errors.New("PORT environment variable not found")
	}
	log.Info("Loaded", "PORT", portStr)

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return Config{}, fmt.Errorf("failed to convert PORT environment variable: %s", err)
	}

	project, ok := os.LookupEnv("PROJECT")
	if !ok {
		return Config{}, errors.New("PROJECT environment variable not found")
	}
	log.Info("Loaded", "PROJECT", project)

	gcloudPollIntervalStr, ok := os.LookupEnv("GCLOUD_POLL_INTERVAL")
	if !ok {
		gcloudPollIntervalStr = "10m"
	}
	log.Info("Loaded", "GCLOUD_POLL_INTERVAL", gcloudPollIntervalStr)

	gcloudPollInterval, err := time.ParseDuration(gcloudPollIntervalStr)
	if err != nil {
		return Config{}, fmt.Errorf("failed to parse GCLOUD_POLL_INTERVAL environment variable: %s", err)
	}

	clusterLifetimeDurationStr, ok := os.LookupEnv("CLUSTER_LIFETIME_DURATION")
	if !ok {
		clusterLifetimeDurationStr = "24h"
	}
	log.Info("Loaded", "CLUSTER_LIFETIME_DURATION", clusterLifetimeDurationStr)

	clusterLifetimeDuration, err := time.ParseDuration(clusterLifetimeDurationStr)
	if err != nil {
		return Config{}, fmt.Errorf("failed to parse CLUSTER_LIFETIME_DURATION environment variable: %s", err)
	}

	return Config{
		Port:                    port,
		Project:                 project,
		GCloudPollInterval:      gcloudPollInterval,
		ClusterLifetimeDuration: clusterLifetimeDuration,
	}, nil
}
