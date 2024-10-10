package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/luthermonson/go-proxmox"
)

type Host struct {
	HostId        int    `json:"id"`
	Name          string `json:"name"`
	MacAddress    string `json:"macaddress"`
	IpAddress     string `json:"ipaddress"`
	AlternatePort string `json:"alternateport"`
	OnlineStatus  bool   `json:"onlinestatus"`
	IsProxmox     bool   `json:"isProxmox"`
}

func (a Host) Value() (driver.Value, error) {
	return json.Marshal(a)
}
func (a *Host) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &a)
}

type PVEHost struct {
	PVEId         int    `json:"id"`
	Name          string `json:"name"`
	MacAddress    string `json:"macAddress"`
	IpAddress     string `json:"ipAddress"`
	AlternatePort string `json:"alternatePort"`
	OnlineStatus  bool   `json:"onlineStatus"`
	Credentials   proxmox.Credentials
	ApiKey        string
}
