package database

import (
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq" // don't forget to add it. It doesn't be added automatically
	"github.com/luthermonson/go-proxmox"
	models "github.com/tophmayor/wakehost/models"
)

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}
type PostgresConfigStore struct {
	Databases map[string]PostgresConfig
}

func (cfg PostgresConfig) String() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)
}

type ProxmoxStore struct {
	ProxmoxClients map[string]*proxmox.Client
}

var Db *sql.DB //created outside to make it global.
var CurrentProxmoxClient *proxmox.Client
var DbConfig PostgresConfigStore
var SelectedConfigName string
var DBConnected bool
var ConfigNeeded bool

func LoadDatabaseConfig() error {
	content, configErr := os.ReadFile("config.json")
	if configErr != nil {
		fmt.Println("Please make sure config.json exists in the same place sample_config.json.")
		return configErr
	}
	jsonErr := json.Unmarshal(content, &DbConfig)
	if jsonErr != nil {
		return jsonErr
	}
	if len(DbConfig.Databases) == 0 {
		fmt.Println("error")
		return fmt.Errorf("nothing in map")
	}
	return nil
}

func ConnectDatabase() {
	if SelectedConfigName == "" {
		SelectedConfigName = "wakehost"
	}
	psqlSetup := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		DbConfig.Databases[SelectedConfigName].Host, DbConfig.Databases[SelectedConfigName].Port, DbConfig.Databases[SelectedConfigName].User, DbConfig.Databases[SelectedConfigName].Name, DbConfig.Databases[SelectedConfigName].Password, DbConfig.Databases[SelectedConfigName].SSLMode)
	db, errSql := sql.Open("postgres", psqlSetup)
	if errSql != nil {
		fmt.Println("There is an error while connecting to the database ", errSql)
		panic(errSql)
	} else {
		Db = db
		DBConnected = true
		fmt.Println("Successfully connected to database!")
	}
}
func ConnectProxmox(pveHost models.PVEHost) {
	insecureHTTPClient := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	fmt.Println("pveHost: ", pveHost)
	credentials := proxmox.Credentials{
		Username: pveHost.Credentials.Username + "@pam",
		Password: pveHost.Credentials.Password,
	}
	fmt.Println("credentials: ", credentials)
	proxmoxClient := proxmox.NewClient(`http://`+pveHost.IpAddress+`:8006/api2/json`,
		proxmox.WithCredentials(&credentials),
		proxmox.WithHTTPClient(&insecureHTTPClient),
	)
	fmt.Println("proxmoxClient: ", proxmoxClient)
	CurrentProxmoxClient = proxmoxClient
}

func PostSetupHandler(c *gin.Context) {
	sslEnabled := c.PostForm("ssl")
	if sslEnabled != "enable" {
		sslEnabled = "disable"
	}
	newConfig := PostgresConfig{Host: c.PostForm("dbip"),
		Port:     c.PostForm("dbport"),
		User:     c.PostForm("dbuser"),
		Password: c.PostForm("dbpassword"),
		Name:     c.PostForm("dbname"),
		SSLMode:  sslEnabled,
	}
	fmt.Println("newConfig:", newConfig)

	DbConfig.Databases[newConfig.Name] = newConfig
	content, _ := json.Marshal(DbConfig)
	writeErr := os.WriteFile("config.json", content, 0644)
	if writeErr != nil {
		log.Fatal(writeErr)
	}
	ConfigNeeded = false
	SelectedConfigName = newConfig.Name
	c.Redirect(302, "/home")
}
