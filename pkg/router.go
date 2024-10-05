package router

import (
	"context"
	"crypto/tls"
	"embed"
	"fmt"
	"html/template"
	io "io/fs"
	"net/http"
	// "database/sql"
	// brotli "github.com/anargu/gin-brotli"
	// "github.com/gin-contrib/sessions"
	// "github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	// csrf "github.com/utrack/gin-csrf"
	"github.com/luthermonson/go-proxmox"
	// _ "github.com/jackc/pgx/v4/stdlib"
)

//go:embed templates/*
var templateFS embed.FS

//go:embed templates/static/*
var staticFiles embed.FS
var hostStore *HostStore
var pvehostStore *PVEHostStore

func Router() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// store := cookie.NewStore([]byte("secret"))
	// r.Use(sessions.Sessions("mysession", store))
	// r.Use(csrf.Middleware(csrf.Options{
	// 	Secret: "secret123",
	// 	ErrorFunc: func(c *gin.Context) {
	// 		c.String(400, "CSRF token mismatch")
	// 		c.Abort()
	// 	},
	// }))
	tmpl := template.Must(template.ParseFS(templateFS, "templates/bootstrap/*"))
	hostStore = new(HostStore)
	pvehostStore = new(PVEHostStore)

	hostStore, _ = loadHosts()
	pvehostStore, _ = loadPVEHosts()
	checkifHostsOnline(hostStore)
	insecureHTTPClient := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	credentials := proxmox.Credentials{
		Username: pvehostStore.Hosts["pve"].Credentials.Username + "@pam",
		Password: pvehostStore.Hosts["pve"].Credentials.Password,
	}
	client := proxmox.NewClient("http://"+pvehostStore.Hosts["pve"].IpAddress+":8006/api2/json",
		proxmox.WithCredentials(&credentials),
		proxmox.WithHTTPClient(&insecureHTTPClient),
	)

	version, err := client.Version(context.Background())
	if err != nil {
		// r.GET("/home", func(c *gin.Context) {
		// 	c.HTML(http.StatusOK, "index.html", gin.H{})
		// })
	}
	// } else {
	// 	pve, _ := client.Node(context.Background(), "pve")
	// 	// fmt.Println(version.Release) // 7.4

	// }
	// 	// fmt.Println(version.Release) // 7.4

	// vm, _ := pve.VirtualMachine(context.Background(), 104)
	// vms, _ := pve.VirtualMachines(context.Background())
	// fmt.Println("name: ", vms)
	// fmt.Println("name: ", vm)

	// vm.Start(context.Background())
	// channel.connected_devices = make(map[string]map[string]StringBool)
	r.SetHTMLTemplate(tmpl)
	// r.Use(sessionMiddleware())
	// r.Use(gin.Recovery())
	// r.Use(brotli.Brotli(brotli.DefaultCompression))
	static := r.Group("/")
	{
		static.Use(cacheMiddleware())
		fs, _ := io.Sub(staticFiles, "templates/static")
		static.StaticFS("/static/", http.FS(fs))
	}
	r.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/home")
	})
	r.GET("/home", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})
	r.GET("/registeredhosts", func(c *gin.Context) {
		checkifHostsOnline(hostStore)
		c.HTML(http.StatusOK, "hosts.html", gin.H{"Hosts": hostStore.Hosts})

	})
	r.GET("/addhost", func(c *gin.Context) {
		c.HTML(http.StatusOK, "addhosts.html", gin.H{"Hosts": hostStore.Hosts})
	})
	r.POST("/addhost", func(c *gin.Context) {
		host := WakeOnLanHost{Name: c.PostForm("name"),
			MacAddress:    c.PostForm("macAddress"),
			IpAddress:     c.PostForm("ipAddress"),
			AlternatePort: c.PostForm("alternatePort"),
		}
		addHost(host, hostStore)
		c.Redirect(302, "/registeredhosts")
	})
	r.GET("/addpvehost", func(c *gin.Context) {
		c.HTML(http.StatusOK, "addpvehost.html", gin.H{"Hosts": pvehostStore.Hosts})
	})
	r.POST("/addpvehost", func(c *gin.Context) {
		// insecureHTTPClient := http.Client{
		// 	Transport: &http.Transport{
		// 		TLSClientConfig: &tls.Config{
		// 			InsecureSkipVerify: true,
		// 		},
		// 	},
		// }
		// credentials := proxmox.Credentials{
		// 	Username: c.PostForm("username"),
		// 	Password: c.PostForm("password"),
		// }
		// client := proxmox.NewClient("http://"+c.PostForm("ipAddress")+":8006/api2/json",
		// 	proxmox.WithCredentials(&credentials),
		// 	proxmox.WithHTTPClient(&insecureHTTPClient),
		// )

		// version, err := client.Version(context.Background())
		host := PVEHost{Name: c.PostForm("name"),
			Credentials: proxmox.Credentials{
				Username: c.PostForm("username"),
				Password: c.PostForm("password"),
			},
			IpAddress: c.PostForm("ipAddress"),
		}
		addPVEHost(host, pvehostStore)
		if err != nil {
			panic(err)
		}
		fmt.Println(version.Release) // 7.4
		c.Redirect(302, "/proxmoxhosts")
	})
	r.GET("/pvehosts/", func(c *gin.Context) {
		nodes, err := client.Nodes(context.Background())
		if err != nil || nodes[0] == nil {
			c.Redirect(302, "/registeredhosts")
		}
		// fmt.Println("nodes:", nodes[0])
		c.HTML(http.StatusOK, "proxmoxhosts.html", gin.H{"Hosts": nodes})
	})
	r.GET("/pvehosts/:name", func(c *gin.Context) {
		name := c.Param("name")
		node, _ := client.Node(context.Background(), name)
		nodeVms, _ := node.VirtualMachines(context.Background())
		fmt.Println("vms:", nodeVms)
		c.HTML(http.StatusOK, "proxmoxhost.html", gin.H{"Hosts": nodeVms})
	})
	r.POST("/registeredhosts", func(c *gin.Context) {
		sendWol(hostStore, c.PostForm("wol"))
		c.Redirect(302, "/registeredhosts")
	})
	r.POST("/pvehosts/:name", func(c *gin.Context) {
		var start bool
		name := c.Param("name")
		pve, _ := client.Node(context.Background(), name)

		strstart := c.PostForm("start")
		fmt.Printf("ip: %s\n", strstart)
		if strstart == "start" {
			start = true
		} else {
			start = false
		}
		useVM(pve, c.PostForm("vm"), start)
		c.Redirect(302, "/pvehosts")
	})
	r.POST("/pvehosts/", func(c *gin.Context) {
		sendWol(hostStore, c.PostForm("wol"))
		c.Redirect(302, "/registeredhosts")
	})
	// r.GET("/files", ListFiles)
	// r.GET("/upload", func(c *gin.Context) {
	// 	c.HTML(http.StatusOK, "upload.html", gin.H{"csrf": csrf.GetToken(c)})
	// })
	// r.POST("/upload", UploadFiles)
	// r.GET("/connected_devices", func(c *gin.Context) {
	// 	c.HTML(http.StatusOK, "connectedDevices.html", gin.H{"devices": channel.connected_devices})
	// })
	// r.GET("/qr", func(c *gin.Context) {
	// 	c.HTML(http.StatusOK, "generateQR.html", gin.H{"qr": generateQR()})
	// })
	// r.GET("/download/:filename", DownloadFile)
	// r.GET("/preview/:filename", PreviewFile)
	// r.GET("/delete/:filename", DeleteFile)
	// r.GET("/downloadAll", DownloadAllFiles)
	// r.GET("/deleteAll", DeleteAllFiles)
	// r.GET("/setpassword", func(c *gin.Context) {
	// 	c.HTML(http.StatusOK, "setPassword.html", gin.H{"csrf": csrf.GetToken(c)})
	// })
	// r.GET("/verifypassword", func(c *gin.Context) {
	// 	c.HTML(http.StatusOK, "verifyPassword.html", gin.H{"csrf": csrf.GetToken(c)})
	// })
	// r.POST("/setpassword", func(c *gin.Context) {
	// 	password := c.PostForm("password")
	// 	setPassword(password)
	// 	c.Redirect(302, "/")
	// })
	// r.POST("/verifypassword", func(c *gin.Context) {
	// 	password := c.PostForm("password")
	// 	iscorrect := verifyPassword(password)
	// 	if iscorrect {
	// 		c.Redirect(302, "/")
	// 	}
	// })

	return r
}
