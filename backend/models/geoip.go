package models

type GeoIPInfo struct {
	Country     string `json:"country,omitempty"`
	CountryCode string `json:"country_code,omitempty"`
	Region      string `json:"region,omitempty"`
	RegionName  string `json:"region_name,omitempty"`
	City        string `json:"city,omitempty"`
	Timezone    string `json:"timezone,omitempty"`
	ISP         string `json:"isp,omitempty"`
	IP          string `json:"ip,omitempty"`
}
