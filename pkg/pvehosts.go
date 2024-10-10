package router

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/luthermonson/go-proxmox"
	database "github.com/tophmayor/wakehost/db"
	models "github.com/tophmayor/wakehost/models"
)

func ContainsPVEHost(hosts map[string]models.PVEHost, host models.PVEHost) bool {
	for k, v := range hosts {
		if k == host.Name && v == host {
			return true
		}
	}
	return false
}
func ComparePVEHosts(host1 models.PVEHost, host2 models.PVEHost) bool {

	if host1.Name == host2.Name {
		return true
	}
	if host1.MacAddress == host2.MacAddress {
		return true
	}
	if host1.IpAddress == host2.IpAddress {
		return true
	}
	return false
}

func getPVEHostHandler(c *gin.Context) {
	if database.ConfigNeeded {
		c.Redirect(302, "/setup")
	} else {
		name := c.Param("name")
		currentPVEHost = pveHosts[name]
		database.ConnectProxmox(currentPVEHost)
		proxClient := database.CurrentProxmoxClient

		node, nodeErr := proxClient.Node(context.Background(), name)
		if nodeErr != nil {
			panic(nodeErr)
		}
		nodeVms, _ := node.VirtualMachines(context.Background())
		c.HTML(http.StatusOK, "proxmoxhost.html", gin.H{"Hosts": nodeVms})

	}
}

func postPVEHostHandler(c *gin.Context) {
	var start bool
	name := c.Param("name")
	proxClient := database.CurrentProxmoxClient
	pve, _ := proxClient.Node(context.Background(), name)

	strstart := c.PostForm("start")
	if strstart == "start" {
		start = true
	} else {
		start = false
	}
	useVM(pve, c.PostForm("vm"), start)
	c.Redirect(302, "/pvehosts/"+name)
}
func addPVEHostHandler(c *gin.Context) {
	newPVEhost := models.PVEHost{Name: c.PostForm("name"),
		MacAddress:    c.PostForm("macAddress"),
		IpAddress:     c.PostForm("ipAddress"),
		AlternatePort: c.PostForm("alternatePort"),
		OnlineStatus:  pingHost((c.PostForm("ipAddress"))),
		Credentials: proxmox.Credentials{
			Username: c.PostForm("username"),
			Password: c.PostForm("password"),
		},
	}
	addPVEHost(newPVEhost)
	c.Redirect(302, "/pvehosts")
}
func addPVEHost(newPVEhost models.PVEHost) {
	rows, rowErr := database.Db.Query(`
	SELECT *
	FROM pve1
	WHERE name=$1 OR  macAddress=$2 OR ipAddress=$3`, newPVEhost.Name, newPVEhost.MacAddress, newPVEhost.IpAddress)
	if rowErr != nil {
		panic(rowErr)
	}
	if rows != nil {
		for rows.Next() {
			var host models.PVEHost
			err := rows.Scan(&newPVEhost.PVEId, &newPVEhost.Name, &newPVEhost.Credentials.Username, &newPVEhost.Credentials.Password, &newPVEhost.MacAddress, &newPVEhost.IpAddress, &newPVEhost.AlternatePort, &newPVEhost.OnlineStatus, &newPVEhost.ApiKey)
			if err != nil {
				panic(err)
			}

			if ComparePVEHosts(host, newPVEhost) {
				break
			}
		}
	}
	if !ContainsPVEHost(pveHosts, newPVEhost) {
		database.Db.Exec(`
		INSERT INTO pve1(name, macAddress, ipAddress, alternatePort, onlineStatus, username, password, apiKey) 
		VALUES($1, $2, $3, $4, $5, $6, $7, $8);
	`, newPVEhost.Name, newPVEhost.MacAddress, newPVEhost.IpAddress, newPVEhost.AlternatePort, newPVEhost.OnlineStatus, newPVEhost.Credentials.Username, newPVEhost.Credentials.Password, "")
	}
}
func getPVEHosts() {
	var currentHosts = map[string]models.PVEHost{}
	rows, _ := database.Db.Query(`
	SELECT *
	FROM pve1`)
	for rows.Next() {
		var host models.PVEHost
		err := rows.Scan(&host.PVEId, &host.Name, &host.Credentials.Username, &host.Credentials.Password, &host.MacAddress, &host.IpAddress, &host.AlternatePort, &host.OnlineStatus, &host.ApiKey)
		if err != nil {
			panic(err)
		}
		currentHosts[host.Name] = host
	}
	rows.Close()
	pveHosts = currentHosts
}
func startVM(node *proxmox.Node, key string) {
	vmid, _ := strconv.Atoi(key)
	vm, _ := node.VirtualMachine(context.Background(), vmid)
	vm.Start(context.Background())
}
func stopVM(node *proxmox.Node, key string) {
	vmid, _ := strconv.Atoi(key)
	vm, _ := node.VirtualMachine(context.Background(), vmid)
	vm.Stop(context.Background())
}
func useVM(node *proxmox.Node, key string, start bool) {
	if start {
		startVM(node, key)
	} else {
		stopVM(node, key)
	}
}
