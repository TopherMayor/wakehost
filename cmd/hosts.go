package wakehostapi

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

func (cfg PostgresConfig) String() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode)
}

type Host struct {
	Name          string `json:"name"`
	MacAddress    string `json:"macAddress"`
	IpAddress     string `json:"ipAddress"`
	AlternatePort string `json:"alternatePort"`
	OnlineStatus  bool   `json:"onlineStatus"`
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

// func getHosts(c *gin.Context){

// }
