package handlers

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	core_v1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"

	"github.com/mayadata-io/kiali/config"
	"github.com/mayadata-io/kiali/kubernetes"
	"github.com/mayadata-io/kiali/log"
	"github.com/mayadata-io/kiali/models"
	"github.com/mayadata-io/kiali/status"
)

type serviceSupplier func(string, string, string) (*core_v1.ServiceSpec, error)
type dashboardSupplier func(string, string, string) ([]byte, int, error)

// GetGrafanaInfo provides the Grafana URL and other info, first by checking if a config exists
// then (if not) by inspecting the Kubernetes Grafana service in namespace istio-system
func GetGrafanaInfo(w http.ResponseWriter, r *http.Request) {
	info, code, err := getGrafanaInfo(getService, findDashboard)
	if err != nil {
		log.Error(err)
		RespondWithError(w, code, err.Error())
		return
	}
	RespondWithJSON(w, code, info)
}

// getGrafanaInfo returns the Grafana URL and other info, the HTTP status code (int) and eventually an error
func getGrafanaInfo(serviceSupplier serviceSupplier, dashboardSupplier dashboardSupplier) (*models.GrafanaInfo, int, error) {
	grafanaConfig := config.Get().ExternalServices.Grafana

	if !grafanaConfig.DisplayLink {
		return nil, http.StatusNoContent, nil
	}

	externalURL, _ := status.DiscoverGrafana()
	if externalURL == "" {
		return nil, http.StatusServiceUnavailable, errors.New("Grafana URL is not set in Kiali configuration")
	}

	// Check if URL is valid
	_, err := validateURL(externalURL)
	if err != nil {
		return nil, http.StatusServiceUnavailable, errors.New("Wrong format for Grafana URL: " + err.Error())
	}

	apiURL := externalURL

	// Find the in-cluster URL to reach Grafana's REST API if properties demand so
	if grafanaConfig.InCluster {
		saToken, err := kubernetes.GetKialiToken()
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		spec, err := serviceSupplier(saToken, grafanaConfig.ServiceNamespace, grafanaConfig.Service)
		if err != nil {
			if k8serr.IsNotFound(err) {
				return nil, http.StatusServiceUnavailable, err
			}
			return nil, http.StatusInternalServerError, err
		}
		if spec != nil && len(spec.Ports) == 0 {
			return nil, http.StatusServiceUnavailable, errors.New("no port found for Grafana service, cannot access in-cluster service")
		}
		if spec != nil && len(spec.Ports) > 1 {
			log.Warning("Several ports found for Grafana service, picking the first one")
		}
		if spec != nil {
			apiURL = fmt.Sprintf("http://%s.%s:%d", grafanaConfig.Service, grafanaConfig.ServiceNamespace, spec.Ports[0].Port)
		}
	}

	credentials, err := buildAuthHeader(grafanaConfig)
	if err != nil {
		log.Warning("Failed to build auth header token: " + err.Error())
	}

	// Call Grafana REST API to get dashboard urls
	serviceDashboardPath, err := getDashboardPath(apiURL, grafanaConfig.ServiceDashboardPattern, credentials, dashboardSupplier)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	workloadDashboardPath, err := getDashboardPath(apiURL, grafanaConfig.WorkloadDashboardPattern, credentials, dashboardSupplier)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	grafanaInfo := models.GrafanaInfo{
		URL:                   externalURL,
		ServiceDashboardPath:  serviceDashboardPath,
		WorkloadDashboardPath: workloadDashboardPath,
		VarNamespace:          grafanaConfig.VarNamespace,
		VarService:            grafanaConfig.VarService,
		VarWorkload:           grafanaConfig.VarWorkload,
	}

	return &grafanaInfo, http.StatusOK, nil
}

func getDashboardPath(url string, searchPattern string, credentials string, dashboardSupplier dashboardSupplier) (string, error) {
	body, code, err := dashboardSupplier(url, searchPattern, credentials)
	if err != nil {
		return "", err
	}
	if code != http.StatusOK {
		// Get error message
		var f map[string]string
		err = json.Unmarshal(body, &f)
		if err != nil {
			return "", fmt.Errorf("unknown error from Grafana (%d)", code)
		}
		message, ok := f["message"]
		if !ok {
			return "", fmt.Errorf("unknown error from Grafana (%d)", code)
		}
		return "", fmt.Errorf("error from Grafana (%d): %s", code, message)
	}

	// Status OK, read dashboards info
	var dashboards []map[string]interface{}
	err = json.Unmarshal(body, &dashboards)
	if err != nil {
		return "", err
	}
	if len(dashboards) == 0 {
		return "", fmt.Errorf("no Grafana dashboard found for search pattern '%s'", searchPattern)
	}
	if len(dashboards) > 1 {
		log.Infof("Several Grafana dashboards found for pattern '%s', picking the first one", searchPattern)
	}
	dashPath, ok := dashboards[0]["url"]
	if !ok {
		return "", fmt.Errorf("URL field not found in Grafana dashboard for search pattern '%s'", searchPattern)
	}
	return dashPath.(string), nil
}

func findDashboard(url, searchPattern string, credentials string) ([]byte, int, error) {
	req, err := http.NewRequest(http.MethodGet, url+"/api/search?query="+searchPattern, nil)
	if err != nil {
		return nil, 0, err
	}
	if credentials != "" {
		req.Header.Add("Authorization", credentials)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return body, resp.StatusCode, err
}

func buildAuthHeader(grafanaConfig config.GrafanaConfig) (string, error) {
	var credHeader string
	if grafanaConfig.APIKey != "" {
		credHeader = "Bearer " + grafanaConfig.APIKey
	} else if grafanaConfig.Username != "" {
		if grafanaConfig.Password == "" {
			return "", fmt.Errorf("grafana username set but no Grafana password provided")
		}
		basicAuth := base64.StdEncoding.EncodeToString([]byte(grafanaConfig.Username + ":" + grafanaConfig.Password))
		credHeader = "Basic " + basicAuth
	}
	return credHeader, nil
}
