package router

import (
	"embed"
	"html/template"
	io "io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
	database "github.com/tophmayor/wakehost/db"
	models "github.com/tophmayor/wakehost/models"
)

//go:embed templates/*
var templateFS embed.FS

//go:embed templates/static/*
var staticFiles embed.FS

var hosts = map[string]models.Host{}
var pveHosts = map[string]models.PVEHost{}
var currentHost models.Host
var currentPVEHost models.PVEHost

func loadInitialHosts() {
	getHosts()
	getPVEHosts()
}

func Router() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	tmpl := template.Must(template.ParseFS(templateFS, "templates/bootstrap/*"))
	loadErr := database.LoadDatabaseConfig()
	if loadErr != nil {
		database.ConfigNeeded = true
	} else {
		database.ConfigNeeded = false
		database.ConnectDatabase()
		loadInitialHosts()

	}
	r.SetHTMLTemplate(tmpl)
	static := r.Group("/")
	{
		fs, _ := io.Sub(staticFiles, "templates/static")
		static.StaticFS("/static/", http.FS(fs))
	}
	r.GET("/", getBaseUrlHandler)
	r.GET("/setup", getSetupHandler)
	r.POST("/setup", database.PostSetupHandler)
	r.GET("/home", getHomeHandler)
	r.GET("/registeredhosts", getHostsHandler)
	r.GET("/addhost", getAddHostHandler)
	r.GET("/registeredhosts/edit/:name", getEditHostHandler)
	r.POST("/registeredhosts/edit/:name", postEditHostHandler)
	r.POST("/addhost", addHostHandler)
	r.GET("/pvehosts/:name", getPVEHostHandler)
	r.POST("/registeredhosts", postHostHandler)
	r.POST("/pvehosts/:name", postPVEHostHandler)

	return r
}

func getHomeHandler(c *gin.Context) {
	if database.ConfigNeeded {
		c.Redirect(302, "/setup")
	} else {
		if !database.DBConnected {
			database.ConnectDatabase()
			loadInitialHosts()
		}
		c.HTML(http.StatusOK, "index.html", gin.H{"Databases": database.DbConfig.Databases})
	}
}
func getSetupHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "setup.html", gin.H{})
}
func getBaseUrlHandler(c *gin.Context) {
	if database.ConfigNeeded {
		c.Redirect(302, "/setup")
	} else {
		c.Redirect(302, "/home")
	}
}
