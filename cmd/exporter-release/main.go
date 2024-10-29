package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/yaml.v2"
)

// Struct to represent server configuration and check interval
type Config struct {
	Server struct {
		Port          int    `yaml:"port"`
		MetricsPath   string `yaml:"metrics_path"`
		CheckInterval string `yaml:"check_interval"`
	} `yaml:"server"`
}

// Struct for repositories and charts configuration in YAML
type ReposAndCharts struct {
	Repositories []struct {
		URL    string   `yaml:"url"`
		Charts []string `yaml:"charts"`
	} `yaml:"repositories"`
}

// Struct to represent Helm's index.yaml entry, including release date
type HelmChartIndex struct {
	Entries map[string][]struct {
		Version string `yaml:"version"`
		Created string `yaml:"created"`
	} `yaml:"entries"`
}

// Metric to expose the release version and release date
var (
	releaseVersionMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "chart_release_version",
			Help: "Release version of a Helm chart",
		},
		[]string{"repo", "chart", "version", "release_date"},
	)
)

// Function to load server configuration from YAML file
func loadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration file: %v", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("error processing configuration file: %v", err)
	}

	return &config, nil
}

// Function to load repositories and charts from a YAML file
func loadReposAndChartsFromYAML(filename string) (*ReposAndCharts, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	var reposAndCharts ReposAndCharts
	err = yaml.Unmarshal(data, &reposAndCharts)
	if err != nil {
		return nil, fmt.Errorf("error processing YAML file: %v", err)
	}

	return &reposAndCharts, nil
}

// Function to format the release date to DD-MM-YYYY
func formatReleaseDate(created string) (string, error) {
	parsedTime, err := time.Parse(time.RFC3339, created)
	if err != nil {
		return "", fmt.Errorf("error parsing date: %v", err)
	}
	return parsedTime.Format("02-01-2006"), nil
}

func main() {
	// Prioritize environment variables for configuration
	metricsPath := os.Getenv("METRICS_PATH")
	metricsPort := os.Getenv("METRICS_PORT")
	checkIntervalStr := os.Getenv("CHECK_INTERVAL")

	// Load configuration file only if required variables are not set in the environment
	var config *Config
	var err error

	if metricsPath == "" || metricsPort == "" || checkIntervalStr == "" {
		configPath := "config/config.yaml"
		config, err = loadConfig(configPath)
		if err != nil {
			log.Fatalf("Error loading configuration file: %v", err)
		}
	}

	// Use values from the config if environment variables are not set
	if metricsPath == "" {
		metricsPath = config.Server.MetricsPath
	}
	if metricsPort == "" {
		metricsPort = fmt.Sprintf("%d", config.Server.Port)
	}
	if checkIntervalStr == "" {
		checkIntervalStr = config.Server.CheckInterval
	}

	checkInterval, err := time.ParseDuration(checkIntervalStr)
	if err != nil {
		log.Fatalf("Error processing check interval: %v", err)
	}

	// Register Prometheus metrics and start the HTTP server
	prometheus.MustRegister(releaseVersionMetric)
	http.Handle(metricsPath, promhttp.Handler())

	go func() {
		address := fmt.Sprintf(":%s", metricsPort)
		log.Printf("Metrics server started on port %s with path %s and check interval %s", metricsPort, metricsPath, checkInterval)
		log.Fatal(http.ListenAndServe(address, nil))
	}()

	// Periodically update metrics based on configured interval
	for {
		time.Sleep(checkInterval)
	}
}