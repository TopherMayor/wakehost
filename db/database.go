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

type PostgresTableSchema struct {
	schemaname  string
	tablename   string
	tableowner  string
	tablespace  any
	hasindexes  bool
	hasrules    bool
	hastriggers bool
	rowsecurity bool
}

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
var InitialSetup bool
var WolhostCreated bool
var PVEhostCreated bool

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
		checkIfTablesCreated()

	}
}
func ConnectProxmox(pveHost models.PVEHost) {
	fmt.Println("proxmoxHost: ", pveHost)

	insecureHTTPClient := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	credentials := proxmox.Credentials{
		Username: pveHost.Credentials.Username + "@pam",
		Password: pveHost.Credentials.Password,
	}
	tokenId := credentials.Username + "!" + pveHost.ApiCredentials.TokenId
	secret := pveHost.ApiCredentials.Secret

	if tokenId != "" && secret != "" {
		CurrentProxmoxClient = proxmox.NewClient(`http://`+pveHost.IpAddress+`:8006/api2/json`,
			proxmox.WithAPIToken(tokenId, secret),
			proxmox.WithHTTPClient(&insecureHTTPClient),
		)
		fmt.Println("proxmoxClient: ", CurrentProxmoxClient)

	} else {
		if pveHost.Credentials.Username != "" && pveHost.Credentials.Password != "" {

			CurrentProxmoxClient = proxmox.NewClient(`http://`+pveHost.IpAddress+`:8006/api2/json`,
				proxmox.WithCredentials(&credentials),
				proxmox.WithHTTPClient(&insecureHTTPClient),
			)
			fmt.Println("proxmoxClient: ", CurrentProxmoxClient)

		} else {
			fmt.Println("Credential Error. Failed Client Initialization")

		}
	}

	// CurrentProxmoxClient = proxmoxClient
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

func checkIfTablesCreated() {
	rows, tableErr := Db.Query(`SELECT *
	FROM pg_catalog.pg_tables
	WHERE schemaname != 'pg_catalog' AND 
		schemaname != 'information_schema';`)
	fmt.Println("tables: ", rows)
	if tableErr != nil {
		fmt.Println("err ", tableErr)
	} else {
		if rows != nil {
			for rows.Next() {
				var table PostgresTableSchema
				err := rows.Scan(&table.schemaname, &table.tablename, &table.tableowner, &table.tablespace, &table.hasindexes, &table.hasrules, &table.hastriggers, &table.rowsecurity)
				if err != nil {
					panic(err)
				}
				fmt.Println("table:", table)
				if table.tablename == "wolhosts" {
					WolhostCreated = true
				}
				if table.tablename == "pvehosts" {
					PVEhostCreated = true
				}

			}
		}
	}
	if !WolhostCreated || !PVEhostCreated {
		createDBTables()
	} else {
		fmt.Println("tables exist")
		testDBTables()
	}
}

func testDBTables() {
	fmt.Println("TESTING")

	//test via insert
	//1) wolhosts

	wolRes, wolErr := Db.Exec(`
	INSERT INTO wolhosts(name, macaddress, ipaddress, alternateport, onlinestatus, proxmox) 
	VALUES('test', 'aa:aa:aa:aa:aa:aa', '1.1.1.1', '', false, false);`)
	if wolErr != nil {
		WolhostCreated = false
		fmt.Println("WOLE: ", wolErr)

	} else {
		fmt.Println("WOL Result: ", wolRes)
		// 	_, delErr := Db.Exec(`DELETE FROM wolhosts
		//  WHERE name=test1;`)
		// 	if delErr != nil {
		// 		fmt.Println("panic")
		// 		// panic(errDel)
		// 	}
	}
	_, delErr := Db.Exec(`DELETE FROM wolhosts
	WHERE name='test';`)
	if delErr != nil {
		fmt.Println("panic: ", delErr)
		// panic(errDel)
	}
	//2) pvehosts
	pveRes, pveErr := Db.Exec(`
	INSERT INTO pvehosts(name, macaddress, ipaddress, alternateport, onlinestatus, username, password, apikey, apitoken) 
	VALUES('test1', 'aa:aa:aa:aa:aa:aa', '1.1.1.1', '', false, 'test', 'password', '', '');
`)
	if pveErr != nil {
		PVEhostCreated = false
		fmt.Println("pveE: ", pveErr)

	} else {
		fmt.Println("PVE Result: ", pveRes)

	}
	_, delPVEErr := Db.Exec(`DELETE FROM pvehosts
	WHERE name='test1';`)
	if delPVEErr != nil {
		fmt.Println("panic: ", delPVEErr)
		// panic(errDel)
	}
	if !WolhostCreated || !PVEhostCreated {
		createDBTables()
	} else {
		fmt.Println("Test Complete")

	}
}

func createDBTables() {
	if !WolhostCreated {
		//Try to Drop Tables to prevent creation errors

		Db.Exec(`DROP TABLE IF EXISTS wolhosts;`)

		// create wolhost table
		fmt.Println("creating wolhost table")

		wolResult, wolError := Db.Exec(`CREATE TABLE wolhosts (
		id SERIAL PRIMARY KEY,
		name TEXT UNIQUE NOT NULL,
		macaddress MACADDR UNIQUE NOT NULL,
		ipaddress INET UNIQUE NOT NULL,
		alternateport TEXT NOT NULL,
		onlinestatus BOOLEAN NOT NULL,
		proxmox BOOLEAN NOT NULL
	  );`)
		if wolError != nil {
			fmt.Println("wolError: ", wolError)
		}
		fmt.Println("wolResult: ", wolResult)

	}
	if !PVEhostCreated {
		//Try to Drop Tables to prevent creation errors

		Db.Exec(`DROP TABLE IF EXISTS pvehosts;`)

		fmt.Println("creating pvehost table")
		pveResult, pveError := Db.Exec(`CREATE TABLE pvehosts (
		id SERIAL PRIMARY KEY,
		name TEXT UNIQUE NOT NULL,
		username TEXT UNIQUE NOT NULL,
		password TEXT UNIQUE NOT NULL,
		macaddress MACADDR UNIQUE NOT NULL,
		ipaddress INET UNIQUE NOT NULL,
		alternateport TEXT NOT NULL,
		onlinestatus BOOLEAN NOT NULL,
		apikey TEXT UNIQUE NOT NULL,
		apitoken TEXT UNIQUE NOT NULL          
	  );`)
		if pveError != nil {
			fmt.Println("wolError: ", pveError)
		}
		fmt.Println("wolResult: ", pveResult)
	}

}
