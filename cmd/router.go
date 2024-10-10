package wakehostapi

import (
	"fmt"
	// "time"
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v4/stdlib"
	cors "github.com/rs/cors/wrapper/gin"
	// "log"
)

var hosts = []Host{}

func Router() *gin.Engine {
	// gin.SetMode(gin.ReleaseMode)
	// r := gin.Default()
	cfg := PostgresConfig{
		Host:     "192.168.254.17",
		Port:     "5432",
		User:     "christopher",
		Password: "Dec202011",
		Database: "wakehost",
		SSLMode:  "disable",
	}
	db, err := sql.Open("pgx", cfg.String())

	if err != nil {
		panic(err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Connected!")
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS WakeOnLanHosts (
		Name TEXT,
		MacAddress TEXT,
		IpAddress TEXT,
		AlternatePort TEXT,
		OnlineStatus BOOLEAN
	  );`)
	if err != nil {
		panic(err)
	}
	fmt.Println("Tables created.")

	if err != nil {
		panic(err)
	}
	r := gin.Default()
	config := cors.Options{
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
	}

	r.Use(cors.New(config))
	// returns hosts from DB
	r.GET("/hosts", func(c *gin.Context) {
		// host := new(Host)
		var currentHosts = []Host{}
		rows, _ := db.Query(`
		SELECT *
		FROM wakeonlanhosts`)
		for rows.Next() {
			var host Host
			err := rows.Scan(&host.Name, &host.MacAddress, &host.IpAddress, &host.AlternatePort, &host.OnlineStatus)
			if err != nil {
				panic(err)
			}
			if !Contains(hosts, host) {
				currentHosts = append(currentHosts, host)
			}
		}
		rows.Close()
		hosts = currentHosts
		c.IndentedJSON(http.StatusOK, currentHosts)
		fmt.Println("hosts:", hosts)
	})
	// r.GET("/pvehosts", getPVEHosts)

	// add host to DB
	r.POST("/hosts", func(c *gin.Context) {

		// host := new(Host)
		newhost := Host{Name: c.PostForm("name"),
			MacAddress:    c.PostForm("macAddress"),
			IpAddress:     c.PostForm("ipAddress"),
			AlternatePort: c.PostForm("alternatePort"),
			OnlineStatus:  false,
		}
		// 	newhost := Host{Name: "pve",
		// 	MacAddress:    "3C:7C:3F:23:5C:C5",
		// 	IpAddress:      "192.168.254.10",
		// 	AlternatePort: "",
		// 	OnlineStatus:false,
		// }
		rows, _ := db.Query(`
		SELECT *
		FROM wakeonlanhosts
		WHERE name=$1 OR  macAddress=$2 OR ipAddress=$3`, newhost.Name, newhost.MacAddress, newhost.IpAddress)

		for rows.Next() {
			var host Host
			err := rows.Scan(&host.Name, &host.MacAddress, &host.IpAddress, &host.AlternatePort, &host.OnlineStatus)
			if err != nil {
				panic(err)
			}

			if Compare(host, newhost) {
				break
			}
		}
		if !Contains(hosts, newhost) {
			_, err = db.Exec(`
			INSERT INTO wakeonlanhosts(name, macAddress, ipAddress, alternatePort, onlineStatus) 
			VALUES($1, $2, $3, $4, $5);
		`, newhost.Name, newhost.MacAddress, newhost.IpAddress, newhost.AlternatePort, newhost.OnlineStatus)
		}

		c.Redirect(302, "/hosts")
	})
	r.PATCH("/hosts", func(c *gin.Context) {
		var columns = []string{"name", "ipAddress"}
		var cmd = ""
		for i := range columns {
			cmd += columns[i] + "='test',"
		}
		fmt.Println("cmd:", cmd)
		cmd = cmd[:len(cmd)-1]
		var final = `UPDATE wakeonlanhosts
		SET ` + cmd + ` WHERE name ='pve';`
		fmt.Println("final:", final)
		_, err := db.Exec(final)
		if err != nil {
			fmt.Println("panic")

			panic(err)
		}
		c.Redirect(302, "/hosts")

		fmt.Println("hosts:", hosts)
	})

	// r.POST("/pvehosts", addPVEHost)
	// r.GET("/hosts/:id", getHostById)
	// r.GET("/pvehosts/:id", getPVEHostById)
	// r.PATCH("/pvehosts/:id", editPVEHostById)
	// r.DELETE("/pvehosts/:id", deletePVEHostById)
	// currentTime := time.Now()
	// url := "http://" + router.GetOutboundIP() + ":8080/"
	// fmt.Println(currentTime.Format("Monday 02-Jan-2006 15:04:00 PM"))
	// fmt.Println("Starting server...")
	// fmt.Println("Listening on : ", url)
	// r := router.Router()
	// router.OpenBrowser(url)
	// r.Run("0.0.0.0:8080")
	// r.Run("0.0.0.0:8080")
	return r
}
