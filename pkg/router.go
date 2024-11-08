package router

import (
	"github.com/gin-gonic/gin"
	database "github.com/tophmayor/wakehost/db"
	models "github.com/tophmayor/wakehost/models"
)

//go:embed templates/*
// var templateFS embed.FS

var hosts = []models.Host{}
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

	// tmpl := template.Must(template.ParseFS(templateFS, "templates/bootstrap/*"))
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
			hosts.PUT("/:id", updateHost)
			hosts.DELETE("/:id", DeleteHost)
		}

		pvehosts := api.Group("/pvehosts")
		{
			pvehosts.GET("", getAllPVEHosts)
			pvehosts.GET("/:id", getPVEHostHandler)
			// pvehosts.POST("", addPVEHost)
			// pvehosts.PUT("/:id", handlers.UpdatePVEHost)
			// pvehosts.DELETE("/:id", handlers.DeletePVEHost)
		}
	}

	// r.SetHTMLTemplate(tmpl)

	// r.GET("/", getBaseUrlHandler)
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

// Handler Functions
// func getHomeHandler(c *gin.Context) {
// 	if database.ConfigNeeded {
// 		c.Redirect(302, "/setup")
// 	} else {
// 		if !database.DBConnected {
// 			database.ConnectDatabase()
// 			loadInitialHosts()
// 		}
// 		c.HTML(http.StatusOK, "index.html", gin.H{"Databases": database.DbConfig.Databases})
// 	}
// }
// func getSetupHandler(c *gin.Context) {
// 	c.HTML(http.StatusOK, "setup.html", gin.H{})
// }
// func getBaseUrlHandler(c *gin.Context) {
// 	if database.ConfigNeeded {
// 		c.Redirect(302, "/setup")
// 	} else {
// 		c.Redirect(302, "/home")
// 	}
// }
