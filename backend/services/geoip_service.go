package services

import (
	"encoding/json"
	"fmt"
	"github.com/andriyg76/bgl/models"
	"github.com/andriyg76/glog"
	"io"
	"net/http"
	"time"
)

type GeoIPService interface {
	GetGeoIPInfo(ipAddress string) (*models.GeoIPInfo, error)
}

type geoipService struct {
	client *http.Client
}

func NewGeoIPService() GeoIPService {
	return &geoipService{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// ipapiResponse matches the response structure from ipapi.co
type ipapiResponse struct {
	IP         string `json:"ip"`
	City       string `json:"city"`
	Region     string `json:"region"`
	RegionCode string `json:"region_code"`
	Country    string `json:"country_name"`
	CountryCode string `json:"country_code"`
	Timezone   string `json:"timezone"`
	Org        string `json:"org"`
	Error      bool   `json:"error"`
	Reason     string `json:"reason"`
}

func (s *geoipService) GetGeoIPInfo(ipAddress string) (*models.GeoIPInfo, error) {
	if ipAddress == "" {
		return nil, fmt.Errorf("ip address is required")
	}

	url := fmt.Sprintf("https://ipapi.co/%s/json/", ipAddress)
	
	resp, err := s.client.Get(url)
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

	var apiResp ipapiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		glog.Warn("Failed to decode geoip response for %s: %v", ipAddress, err)
		return nil, err
	}

	if apiResp.Error {
		glog.Warn("GeoIP API error for %s: %s", ipAddress, apiResp.Reason)
		return nil, fmt.Errorf("geoip API error: %s", apiResp.Reason)
	}

	return &models.GeoIPInfo{
		IP:          apiResp.IP,
		Country:     apiResp.Country,
		CountryCode: apiResp.CountryCode,
		Region:      apiResp.RegionCode,
		RegionName:  apiResp.Region,
		City:        apiResp.City,
		Timezone:    apiResp.Timezone,
		ISP:         apiResp.Org,
	}, nil
}
