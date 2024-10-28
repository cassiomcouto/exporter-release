package main

import (
	"fmt"
	"io"
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

	// Format the date to DD-MM-YYYY
	return parsedTime.Format("02-01-2006"), nil
}

// Function to get the latest release version and formatted release date of a Helm chart
func getLatestHelmChartRelease(repoURL, chartName string) (string, string, error) {
	resp, err := http.Get(repoURL + "/index.yaml")
	if err != nil {
		return "", "", fmt.Errorf("error accessing Helm repository: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("error accessing Helm repository: status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("error reading index.yaml content: %v", err)
	}

	var index HelmChartIndex
	err = yaml.Unmarshal(body, &index)
	if err != nil {
		return "", "", fmt.Errorf("error processing index.yaml: %v", err)
	}

	// Check if the chart is listed in the index.yaml
	chart, exists := index.Entries[chartName]
	if !exists || len(chart) == 0 {
		return "", "", fmt.Errorf("chart %s not found in repository %s", chartName, repoURL)
	}

	// Format the release date
	formattedDate, err := formatReleaseDate(chart[0].Created)
	if err != nil {
		return "", "", err
	}

	// Return the most recent version and formatted release date
	return chart[0].Version, formattedDate, nil
}

// Function to check the latest release version and update the metric
func checkReleases(reposAndCharts *ReposAndCharts) {
	for _, repo := range reposAndCharts.Repositories {
		for _, chartName := range repo.Charts {
			latestVersion, formattedDate, err := getLatestHelmChartRelease(repo.URL, chartName)
			if err != nil {
				log.Printf("Error getting the release version for chart %s in repository %s: %v\n", chartName, repo.URL, err)
				continue
			}

			// Set the metric with repository, chart name, version, and formatted release date labels
			releaseVersionMetric.With(prometheus.Labels{
				"repo":         repo.URL,
				"chart":        chartName,
				"version":      latestVersion,
				"release_date": formattedDate,
			}).Set(1)
		}
	}
}

func main() {
	// Define the base configuration directory, using CONFIG_PATH or a default value
	configDir := os.Getenv("CONFIG_PATH")
	if configDir == "" {
		if len(os.Args) > 1 {
			configDir = os.Args[1]
		} else {
			configDir = "config"
		}
	}

	// Full paths for configuration files
	configPath := fmt.Sprintf("%s/config.yaml", configDir)
	reposPath := fmt.Sprintf("%s/repos_and_charts.yaml", configDir)

	// Load server configuration from config.yaml
	config, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("Error loading configuration file: %v", err)
	}

	// Get Metrics Path and Port settings, prioritizing environment variables
	metricsPath := os.Getenv("METRICS_PATH")
	if metricsPath == "" {
		metricsPath = config.Server.MetricsPath
	}

	metricsPort := os.Getenv("METRICS_PORT")
	if metricsPort == "" {
		metricsPort = fmt.Sprintf("%d", config.Server.Port)
	}

	// Parse the check interval duration, prioritizing the CHECK_INTERVAL environment variable
	checkIntervalStr := os.Getenv("CHECK_INTERVAL")
	if checkIntervalStr == "" {
		checkIntervalStr = config.Server.CheckInterval
	}

	checkInterval, err := time.ParseDuration(checkIntervalStr)
	if err != nil {
		log.Fatalf("Error processing check interval: %v", err)
	}

	// Register the metric in Prometheus
	prometheus.MustRegister(releaseVersionMetric)

	// Start the HTTP server with port and metrics path defined in the configuration file
	http.Handle(metricsPath, promhttp.Handler())
	go func() {
		address := fmt.Sprintf(":%s", metricsPort)
		log.Printf("Metrics server started on port %s with path %s and check interval %s", metricsPort, metricsPath, checkInterval)
		log.Fatal(http.ListenAndServe(address, nil))
	}()

	// Load the list of repositories and charts from the YAML file
	reposAndCharts, err := loadReposAndChartsFromYAML(reposPath)
	if err != nil {
		log.Fatalf("Error loading repos_and_charts.yaml file: %v", err)
	}

	// Periodically check releases and update metrics
	for {
		checkReleases(reposAndCharts)
		time.Sleep(checkInterval) // Use the check interval from the configuration or environment variable
	}
}