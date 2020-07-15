package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/go-logr/logr"
)

type Config struct {
	Port                    int
	Project                 string
	GCloudPollInterval      time.Duration
	GCloudGKELabelFilters   []string
	ClusterLifetimeDuration time.Duration
	VCAPServices            VCAPServices
	BasicAuthUsername       string
	BasicAuthPassword       string
}

type VCAPServices struct {
	UserProvided []UserProvidedVCAPServices `json:"user-provided"`
}

type UserProvidedVCAPServices struct {
	BindingName    string            `json:"binding_name"`
	Credentials    map[string]string `json:"credentials"`
	InstanceName   string            `json:"instance_name"`
	Label          string            `json:"label"`
	Name           string            `json:"name"`
	SyslogDrainURL string            `json:"syslog_drain_url"`
	Tags           []string          `json:"tags"`
	VolumeMounts   []string          `json:"volume_mounts"`
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

	gcpServiceAccountKey, ok := os.LookupEnv("GCP_SERVICE_ACCOUNT_KEY")
	if !ok {
		return Config{}, errors.New("GCP_SERVICE_ACCOUNT_KEY environment variable not found")
	}
	gcpKeyFile, err := ioutil.TempFile("", "gcp-service-account-key")
	if err != nil {
		return Config{}, fmt.Errorf("failed to create temporary file for gcp service account key: %s", err)
	}
	defer gcpKeyFile.Close()

	_, err = gcpKeyFile.Write([]byte(gcpServiceAccountKey))
	if err != nil {
		return Config{}, fmt.Errorf("failed to write to temporary file for gcp service account key: %s", err)
	}

	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", gcpKeyFile.Name())
	log.Info("Loaded", "GCP_SERVICE_ACCOUNT_KEY", "<redacted>")

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

	var gcloudGKELabelFilter []string
	gcloudGKELabelFilterStr, ok := os.LookupEnv("GCLOUD_GKE_LABEL_FILTERS")
	if ok {
		err = json.Unmarshal([]byte(gcloudGKELabelFilterStr), &gcloudGKELabelFilter)
		if err != nil {
			return Config{}, fmt.Errorf("failed to parse GCLOUD_GKE_LABEL_FILTERS environment variable: %s", err)
		}
		log.Info("Loaded", "GCLOUD_GKE_LABEL_FILTERS", gcloudGKELabelFilter)
	} else {
		log.Info("GCLOUD_GKE_LABEL_FILTERS unset.")
	}

	vcapServicesStr, ok := os.LookupEnv("VCAP_SERVICES")
	if !ok {
		return Config{}, fmt.Errorf("VCAP_SERVICES environment variable not found")
	}
	log.Info("Loaded", "VCAP_SERVICES", "<redacted>")

	var vcapServices VCAPServices
	err = json.Unmarshal([]byte(vcapServicesStr), &vcapServices)
	if err != nil {
		return Config{}, fmt.Errorf("failed to parse VCAP_SERVICES environment variable: %s", err)
	}

	basicAuthUsername, ok := os.LookupEnv("BASIC_AUTH_USERNAME")
	if !ok {
		return Config{}, fmt.Errorf("BASIC_AUTH_USERNAME environment variable not found")
	}
	log.Info("Loaded", "BASIC_AUTH_USERNAME", "<redacted>")

	basicAuthPassword, ok := os.LookupEnv("BASIC_AUTH_PASSWORD")
	if !ok {
		return Config{}, fmt.Errorf("BASIC_AUTH_PASSWORD environment variable not found")
	}
	log.Info("Loaded", "BASIC_AUTH_PASSWORD", "<redacted>")

	return Config{
		Port:                    port,
		Project:                 project,
		GCloudPollInterval:      gcloudPollInterval,
		GCloudGKELabelFilters:   gcloudGKELabelFilter,
		ClusterLifetimeDuration: clusterLifetimeDuration,
		VCAPServices:            vcapServices,
		BasicAuthUsername:       basicAuthUsername,
		BasicAuthPassword:       basicAuthPassword,
	}, nil
}

func (c Config) GetUserProvidedVCAPServiceByBindingName(bindingName string) (UserProvidedVCAPServices, bool) {
	for _, service := range c.VCAPServices.UserProvided {
		if service.BindingName == bindingName {
			return service, true
		}
	}

	return UserProvidedVCAPServices{}, false
}
