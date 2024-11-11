package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	database "github.com/tophmayor/wakehost/db"
	models "github.com/tophmayor/wakehost/models"
)

var wolHosts = []models.Host{}
var pveHosts = []models.PVEHost{}
var currentHost models.Host
var currentPVEHost models.PVEHost
var CurrentDB database.PostgresConfig

func loadInitialHosts() {
	getHosts()
	getPVEHosts()
}

func Router() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	config := cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}
	r.Use(cors.New(config))
	loadErr := database.LoadDatabaseConfig()
	if loadErr != nil {
		database.ConfigNeeded = true
	} else {
		database.ConfigNeeded = false
		database.ConnectDatabase()
		CurrentDB = database.DbConfig.Databases[database.SelectedConfigName]

		loadInitialHosts()

	}
	api := r.Group("/api")
	{
		hosts := api.Group("/hosts")
		{
			hosts.GET("", getHostsHandler)
			hosts.GET("/:id", getHostByIDHandler)
			hosts.POST("", addHostHandler)
			hosts.PUT("/:id", updateHostHandler)
			hosts.DELETE("/:id", DeleteHost)
		}

		pvehosts := api.Group("/pvehosts")
		{
			pvehosts.GET("", getAllPVEHosts)
			pvehosts.GET("/:id", getPVEHostHandler)
			// pvehosts.POST("", addPVEHost)
			pvehosts.PUT("/:id", pveHostActionHandler)
			// pvehosts.DELETE("/:id", handlers.DeletePVEHost)
		}
	}
	// r.GET("/setup", getSetupHandler)
	// r.POST("/setup", database.PostSetupHandler)
	// r.GET("/home", getHomeHandler)
	// r.POST("/home", func(c *gin.Context) {
	// 	db := c.PostForm("db")
	// 	fmt.Println("db: ", db)
	// 	if db != "" {
	// 		CurrentDB = database.DbConfig.Databases[db]
	// 		c.Redirect(302, "/databases/edit/"+db)
	// 	} else {
	// 		c.Redirect(302, "/home")
	// 	}
	// })

	// r.GET("/registeredhosts", getHostsHandler)
	// r.GET("/addhost", getAddHostHandler)
	// r.GET("/registeredhosts/edit/:name", getEditHostHandler)
	// r.POST("/registeredhosts/edit/:name", postEditHostHandler)
	// r.POST("/addhost", addHostHandler)
	// r.GET("/pvehosts/:name", getPVEHostHandler)
	// r.GET("/pvehosts", getAllPVEHosts)
	// r.POST("/registeredhosts", postHostHandler)
	// r.POST("/pvehosts/:name", postPVEHostHandler)
	// r.GET("/databases/edit/:name", func(c *gin.Context) {
	// 	if database.ConfigNeeded {
	// 		c.Redirect(302, "/setup")
	// 	} else {
	// 		c.HTML(http.StatusOK, "editdb.html", gin.H{"DB": CurrentDB})
	// 	}
	// })
	// r.POST("/databases/edit/:name", func(c *gin.Context) {
	// 	ipadd := c.PostForm("ipAddress")
	// 	port := c.PostForm("port")
	// 	user := c.PostForm("user")
	// 	name := c.PostForm("name")
	// 	password := c.PostForm("password")
	// 	ssl := c.PostForm("ssl")
	// 	if ssl == "" {
	// 		ssl = "disable"
	// 	}

	// 	updatedDB := database.PostgresConfig{
	// 		Host:     ipadd,
	// 		Port:     port,
	// 		User:     user,
	// 		Name:     name,
	// 		Password: password,
	// 		SSLMode:  ssl,
	// 	}
	// 	database.DbConfig.Databases[updatedDB.Name] = updatedDB
	// 	database.SelectedConfigName = updatedDB.Name
	// 	database.ConnectDatabase()

	// 	c.Redirect(302, "/home")

	// })

	return r
}
