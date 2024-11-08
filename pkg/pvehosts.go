package router

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	database "github.com/tophmayor/wakehost/db"
	models "github.com/tophmayor/wakehost/models"
)

// Handler Functions
func getPVEHostHandler(c *gin.Context) {
	getPVEHosts()
	if database.ConfigNeeded {
		c.Redirect(302, "/setup")
	} else {
		idParam := c.Param("id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid host ID"})
			return
		}
		var host models.PVEHost
		err = database.Db.QueryRow("SELECT id, name, macaddress, ipaddress, alternateport, onlinestatus, username, password, apikey, apitoken  FROM pvehosts WHERE id=$1", id).
			Scan(&host.PVEId, &host.Name, &host.MacAddress, &host.IpAddress, &host.AlternatePort, &host.OnlineStatus, &host.Credentials.Username, &host.Credentials.Password, &host.ApiCredentials.Secret, &host.ApiCredentials.TokenId)
		if err != nil {
			fmt.Println("err:", err)
			c.JSON(http.StatusNotFound, gin.H{"error": "Host not found"})
			return
		}
		currentPVEHost = host
		c.JSON(http.StatusOK, host)
		database.ConnectProxmox(currentPVEHost)
	}
}

func getAllPVEHosts(c *gin.Context) {
	getPVEHosts()
	c.JSON(http.StatusOK, pveHosts)
}

// func postPVEHostHandler(c *gin.Context) {
// 	// var start bool
// 	name := c.Param("name")
// 	proxClient := database.CurrentProxmoxClient
// 	pve, _ := proxClient.Node(context.Background(), name)

// 	start := c.PostForm("start")
// 	stop := c.PostForm("stop")
// 	restart := c.PostForm("restart")

// 	fmt.Println("START: ", start)
// 	fmt.Println("START: ", stop)
// 	fmt.Println("START: ", restart)
// 	fmt.Println("START: ", name)

// 	if start != "" {
// 		fmt.Println("START")
// 		go startVM(pve, start)
// 	}
// 	if stop != "" {
// 		fmt.Println("STOP")
// 		go stopVM(pve, stop)
// 	}
// 	if restart != "" {
// 		fmt.Println("RESTART")
// 		go restartVM(pve, restart)
// 	}
// 	// useVM(pve, c.PostForm("vm"), start)
// 	c.Redirect(302, "/pvehosts/"+name)
// }

// Handler Utils
func addPVEHost(newPVEhost models.PVEHost) {
	fmt.Println("newPVEHost: ", newPVEhost)

	rows, rowErr := database.Db.Query(`
	SELECT *
	FROM pvehosts
	WHERE name=$1 OR  macaddress=$2 OR ipaddress=$3`, newPVEhost.Name, newPVEhost.MacAddress, newPVEhost.IpAddress)
	if rowErr != nil {
		panic(rowErr)
	}
	if rows != nil {
		fmt.Println("scanning")

		for rows.Next() {
			var host models.PVEHost
			err := rows.Scan(&newPVEhost.PVEId, &newPVEhost.Name, &newPVEhost.Credentials.Username, &newPVEhost.Credentials.Password, &newPVEhost.MacAddress, &newPVEhost.IpAddress, &newPVEhost.AlternatePort, &newPVEhost.OnlineStatus, &newPVEhost.ApiCredentials.Secret, &newPVEhost.ApiCredentials.TokenId)
			if err != nil {
				panic(err)
			}

			if ComparePVEHosts(host, newPVEhost) {
				break
			}
		}
	}
	if !ContainsPVEHost(pveHosts, newPVEhost) {
		fmt.Println("inserting")

		_, pveErr := database.Db.Exec(`
		INSERT INTO pvehosts(name, macaddress, ipaddress, alternateport, onlinestatus, username, password, apikey, apitoken) 
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9);
	`, newPVEhost.Name, newPVEhost.MacAddress, newPVEhost.IpAddress, newPVEhost.AlternatePort, newPVEhost.OnlineStatus, newPVEhost.Credentials.Username, newPVEhost.Credentials.Password, newPVEhost.ApiCredentials.Secret, newPVEhost.ApiCredentials.TokenId)
		if pveErr != nil {
			fmt.Println("pveErrInsert: ", pveErr)
		}
		pveHosts = append(pveHosts, newPVEhost)
	}
}

// func findPVEColumnDiffs(oldData models.PVEHost, newData models.PVEHost) ([]string, []string) {
// 	var columns []string
// 	var values []string

// 	if oldData.Name != newData.Name {
// 		columns = append(columns, "name")
// 		values = append(values, newData.Name)
// 	}
// 	if oldData.IpAddress != newData.IpAddress {
// 		columns = append(columns, "ipAddress")
// 		values = append(values, newData.IpAddress)
// 	}
// 	if oldData.MacAddress != newData.MacAddress {
// 		columns = append(columns, "macAddress")
// 		values = append(values, newData.MacAddress)
// 	}
// 	if oldData.AlternatePort != newData.AlternatePort {
// 		columns = append(columns, "alternatePort")
// 		alternatePort := newData.AlternatePort
// 		values = append(values, alternatePort)
// 	}
// 	if oldData.Credentials != newData.Credentials && newData.Credentials.Password != "" {
// 		columns = append(columns, "username")
// 		username := newData.Credentials.Username
// 		values = append(values, username)
// 		columns = append(columns, "password")
// 		password := newData.Credentials.Password
// 		values = append(values, password)

// 	}
// 	if oldData.ApiCredentials != newData.ApiCredentials && newData.ApiCredentials.Secret != "" {
// 		columns = append(columns, "apikey")
// 		secret := newData.ApiCredentials.Secret
// 		values = append(values, secret)
// 		columns = append(columns, "apitoken")
// 		token := newData.ApiCredentials.TokenId
// 		values = append(values, token)
// 	}
// 	return columns, values
// }
// func updatePVEHost(host models.PVEHost) {
// 	columns, values := findPVEColumnDiffs(currentPVEHost, host)

// 	cmd := ""
// 	for i, v := range columns {
// 		if v == "macAddress" || v == "ipAddress" {
// 			if v == "macAddress" {
// 				cmd += v + "=" + `CAST('` + values[i] + `' AS macaddr)`

// 			} else {
// 				cmd += v + "=" + `CAST('` + values[i] + `' AS inet)`
// 			}
// 		} else {
// 			cmd += v + `='` + values[i] + `'`
// 		}
// 		if i < len(columns)-1 {
// 			cmd += ", "
// 		}
// 	}
// 	var final = `UPDATE pvehosts
// 	SET ` + cmd + ` WHERE name='` + currentHost.Name + `';`
// 	_, err := database.Db.Exec(final)
// 	if err != nil {
// 		fmt.Println("panic")

//			panic(err)
//		}
//	}
func getPVEHosts() {
	var currentHosts = []models.PVEHost{}
	rows, rowErr := database.Db.Query(`
	SELECT *
	FROM pvehosts`)
	if rowErr != nil {
		panic(rowErr)
	}
	if rows != nil {
		for rows.Next() {
			var host models.PVEHost
			err := rows.Scan(&host.PVEId, &host.Name, &host.Credentials.Username, &host.Credentials.Password, &host.MacAddress, &host.IpAddress, &host.AlternatePort, &host.OnlineStatus, &host.ApiCredentials.Secret, &host.ApiCredentials.TokenId)
			if err != nil {
				panic(err)
			}
			currentHosts = append(currentHosts, host)
		}
	}
	rows.Close()
	pveHosts = currentHosts
}
func ContainsPVEHost(hosts []models.PVEHost, host models.PVEHost) bool {
	for v := range hosts {
		if hosts[v].Name == host.Name {
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

// Proxmox Functions
// func startVM(node *proxmox.Node, key string) {
// 	vmid, _ := strconv.Atoi(key)
// 	vm, _ := node.VirtualMachine(context.Background(), vmid)
// 	vm.Start(context.Background())
// }
// func stopVM(node *proxmox.Node, key string) {
// 	vmid, _ := strconv.Atoi(key)
// 	vm, _ := node.VirtualMachine(context.Background(), vmid)
// 	vm.Stop(context.Background())
// }
// func restartVM(node *proxmox.Node, key string) {
// 	vmid, _ := strconv.Atoi(key)
// 	vm, _ := node.VirtualMachine(context.Background(), vmid)
// 	vm.Reboot(context.Background())
// }

// func useVM(node *proxmox.Node, key string, start bool) {
// 	if start {
// 		startVM(node, key)
// 	} else {
// 		stopVM(node, key)
// 	}
// }
