package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"time"
)

type OCSResponse struct {
	OCS struct {
		Meta MetaInfo `json:"meta"`
		Data DataInfo `json:"data"`
	} `json:"ocs"`
}

type MetaInfo struct {
	Status     string `json:"status"`
	StatusCode int    `json:"statuscode"`
	Message    string `json:"message"`
}

type DataInfo struct {
	Nextcloud   NextcloudInfo   `json:"nextcloud"`
	Server      ServerInfo      `json:"server"`
	ActiveUsers ActiveUsersInfo `json:"activeUsers"`
}

type NextcloudInfo struct {
	System  NextcloudSystem  `json:"system"`
	Storage NextcloudStorage `json:"storage"`
	Shares  NextcloudShares  `json:"shares"`
}

type NextcloudSystem struct {
	Version   string        `json:"version"`
	Cpuload   []float64     `json:"cpuload"`
	MemTotal  int64         `json:"mem_total"`
	MemFree   int64         `json:"mem_free"`
	SwapTotal int64         `json:"swap_total"`
	SwapFree  int64         `json:"swap_free"`
	Apps      NextcloudApps `json:"apps"`
}

type NextcloudApps struct {
	NumInstalled        int `json:"num_installed"`
	NumUpdatesAvailable int `json:"num_updates_available"`
}

type NextcloudStorage struct {
	NumUsers int `json:"num_users"`
	NumFiles int `json:"num_files"`
}

type NextcloudShares struct {
	NumShares int `json:"num_shares"`
}

type ServerInfo struct {
	PHP      PHPInfo      `json:"php"`
	Database DatabaseInfo `json:"database"`
}

type PHPInfo struct {
	Version string         `json:"version"`
	Opcache PHPOpcacheInfo `json:"opcache"`
}

type PHPOpcacheInfo struct {
	OpcacheStatistics OpcacheStatisticsInfo `json:"opcache_statistics"`
}

type OpcacheStatisticsInfo struct {
	OpcacheHitRate float64 `json:"opcache_hit_rate"`
}

type DatabaseInfo struct {
	Version string `json:"version"`
}

type ActiveUsersInfo struct {
	Last5minutes int `json:"last5minutes"`
}

func checkNextcloud(serverURL string, ncToken string) {
	apiURL := fmt.Sprintf("%s/ocs/v2.php/apps/serverinfo/api/v1/info?format=json&skipApps=false&skipUpdate=false", serverURL)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		fmt.Printf("CRITICAL - Failed to create request: %v\n", err)
		os.Exit(2)
	}
	req.Header.Set("NC-Token", ncToken)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("CRITICAL - API request failed: %v\n", err)
		os.Exit(2)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("CRITICAL - Failed to close response body: %v\n", err)
			os.Exit(2)
		}
	}(resp.Body)

	if resp.StatusCode == http.StatusUnauthorized {
		fmt.Println("CRITICAL - Unauthorized access (401)")
		os.Exit(2)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("CRITICAL - Failed to read API response: %v\n", err)
		os.Exit(2)
	}

	var ocsResp OCSResponse
	err = json.Unmarshal(body, &ocsResp)
	if err != nil {
		fmt.Printf("CRITICAL - Failed to parse API response: %v\n", err)
		os.Exit(2)
	}

	if ocsResp.OCS.Data.Nextcloud.System.Version == "" {
		fmt.Println("CRITICAL - Invalid API response")
		os.Exit(2)
	}

	status := "OK"
	exitCode := 0

	sysInfo := ocsResp.OCS.Data.Nextcloud.System

	if len(sysInfo.Cpuload) >= 3 {
		if sysInfo.Cpuload[0] > 5 || sysInfo.Cpuload[1] > 4 || sysInfo.Cpuload[2] > 3 {
			status = "WARNING - High CPU Load"
			if exitCode < 1 {
				exitCode = 1
			}
		}
	}

	memTotal := sysInfo.MemTotal
	memFree := sysInfo.MemFree
	memUsage := 0.0
	if memTotal > 0 {
		memUsage = (float64(memTotal-memFree) / float64(memTotal)) * 100
	}
	if memUsage > 90 {
		status = "CRITICAL - High Memory Usage"
		if exitCode < 2 {
			exitCode = 2
		}
	} else if memUsage > 80 {
		status = "WARNING - High Memory Usage"
		if exitCode < 1 {
			exitCode = 1
		}
	}

	swapTotal := sysInfo.SwapTotal
	swapFree := sysInfo.SwapFree
	swapUsage := 0.0
	if swapTotal > 0 {
		swapUsage = (float64(swapTotal-swapFree) / float64(swapTotal)) * 100
	}
	if swapUsage > 90 {
		status = "CRITICAL - High Swap Usage"
		if exitCode < 2 {
			exitCode = 2
		}
	} else if swapUsage > 80 {
		status = "WARNING - High Swap Usage"
		if exitCode < 1 {
			exitCode = 1
		}
	}

	if sysInfo.Apps.NumUpdatesAvailable > 0 {
		status = "WARNING - App Updates Available"
		if exitCode < 1 {
			exitCode = 1
		}
	}

	metrics := map[string]interface{}{
		"version":                   sysInfo.Version,
		"num_users":                 ocsResp.OCS.Data.Nextcloud.Storage.NumUsers,
		"num_files":                 ocsResp.OCS.Data.Nextcloud.Storage.NumFiles,
		"cpu_load_1m":               sysInfo.Cpuload[0],
		"cpu_load_5m":               sysInfo.Cpuload[1],
		"cpu_load_15m":              sysInfo.Cpuload[2],
		"memory_total":              memTotal,
		"memory_free":               memFree,
		"memory_usage_percent":      math.Round(memUsage*100) / 100,
		"swap_total":                swapTotal,
		"swap_free":                 swapFree,
		"swap_usage_percent":        math.Round(swapUsage*100) / 100,
		"num_apps_installed":        sysInfo.Apps.NumInstalled,
		"num_apps_update_available": sysInfo.Apps.NumUpdatesAvailable,
		"num_shares":                ocsResp.OCS.Data.Nextcloud.Shares.NumShares,
		"php_version":               ocsResp.OCS.Data.Server.PHP.Version,
		"db_version":                ocsResp.OCS.Data.Server.Database.Version,
		"active_users_5m":           ocsResp.OCS.Data.ActiveUsers.Last5minutes,
		"opcache_hit_rate":          ocsResp.OCS.Data.Server.PHP.Opcache.OpcacheStatistics.OpcacheHitRate,
	}

	metricsOutput := " |"
	for key, value := range metrics {
		metricsOutput += fmt.Sprintf(" %s=%v", key, value)
	}

	fmt.Printf("%s - Nextcloud %s running.%s\n", status, sysInfo.Version, metricsOutput)
	os.Exit(exitCode)
}

func main() {
	server := flag.String("s", "", "Nextcloud Server URL (e.g. https://nextcloud.example.com)")
	token := flag.String("t", "", "Nextcloud NC-Token for API access")

	flag.Parse()

	if *server == "" || *token == "" {
		fmt.Println("CRITICAL - Missing required arguments")
		flag.Usage()
		os.Exit(2)
	}

	checkNextcloud(*server, *token)
}
