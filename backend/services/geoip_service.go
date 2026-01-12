package services

import (
	"encoding/json"
	"fmt"
	"github.com/andriyg76/bgl/models"
	"github.com/andriyg76/glog"
	"io"
	"net/http"
	"os"
	"time"
)

type GeoIPService interface {
	GetGeoIPInfo(ipAddress string) (*models.GeoIPInfo, error)
}

type geoipService struct {
	client   *http.Client
	apiToken string
}

func NewGeoIPService() GeoIPService {
	token := os.Getenv("IPINFO_TOKEN")
	if token == "" {
		glog.Warn("IPINFO_TOKEN is not set, GeoIP lookups will be disabled")
	}
	return &geoipService{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		apiToken: token,
	}
}

// ipinfoResponse matches the response structure from ipinfo.io
type ipinfoResponse struct {
	IP       string `json:"ip"`
	Hostname string `json:"hostname"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Country  string `json:"country"`
	Loc      string `json:"loc"`
	Org      string `json:"org"`
	Postal   string `json:"postal"`
	Timezone string `json:"timezone"`
	Anycast  bool   `json:"anycast"`
	// Error response fields
	Error *ipinfoError `json:"error,omitempty"`
}

type ipinfoError struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

func (s *geoipService) GetGeoIPInfo(ipAddress string) (*models.GeoIPInfo, error) {
	if ipAddress == "" {
		return nil, fmt.Errorf("ip address is required")
	}

	if s.apiToken == "" {
		return nil, fmt.Errorf("IPINFO_TOKEN is not configured")
	}

	url := fmt.Sprintf("https://ipinfo.io/%s", ipAddress)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.apiToken)

	resp, err := s.client.Do(req)
	if err != nil {
		glog.Warn("Failed to fetch geoip info for %s: %v", ipAddress, err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		glog.Warn("GeoIP API returned status %d for %s: %s", resp.StatusCode, ipAddress, string(body))
		return nil, fmt.Errorf("geoip API returned status %d", resp.StatusCode)
	}

	var apiResp ipinfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		glog.Warn("Failed to decode geoip response for %s: %v", ipAddress, err)
		return nil, err
	}

	if apiResp.Error != nil {
		glog.Warn("GeoIP API error for %s: %s - %s", ipAddress, apiResp.Error.Title, apiResp.Error.Message)
		return nil, fmt.Errorf("geoip API error: %s", apiResp.Error.Message)
	}

	return &models.GeoIPInfo{
		IP:          apiResp.IP,
		Country:     apiResp.Country, // ipinfo returns country code (e.g., "US")
		CountryCode: apiResp.Country,
		Region:      apiResp.Region, // ipinfo returns full region name
		RegionName:  apiResp.Region,
		City:        apiResp.City,
		Timezone:    apiResp.Timezone,
		ISP:         apiResp.Org,
	}, nil
}
