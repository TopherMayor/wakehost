package router

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/luthermonson/go-proxmox"
	"github.com/sabhiram/go-wol/wol"
	"github.com/tatsushid/go-fastping"
	database "github.com/tophmayor/wakehost/db"
	models "github.com/tophmayor/wakehost/models"
)

// Handler Functions

// func getEditHostHandler(c *gin.Context) {
// 	if database.ConfigNeeded {
// 		c.Redirect(302, "/setup")
// 	} else {
// 		c.HTML(http.StatusOK, "edithosts.html", gin.H{"Host": currentHost, "PVEHost": currentPVEHost})
// 	}
// }

func getHostsHandler(c *gin.Context) {
	if database.ConfigNeeded {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "data": "", "message": "Failed"})

	} else {
		getHosts()
		// currentTime := time.Now().Format(time.RFC850)
		c.JSON(http.StatusOK, gin.H{"success": true, "data": wolHosts, "message": "Success"})
		fmt.Println("hosts:", wolHosts)
	}
}

// func postHostHandler(c *gin.Context) {
// 	idParam := c.Param("id")
// 	actionParam := c.Param("action")

// 	// wol := c.PostForm("wol")
// 	// del := c.PostForm("delete")
// 	// update := c.PostForm("update")
// 	// addhost := c.PostForm("addhost")

// 	// if del != "" {
// 	// 	fmt.Println("del:", del)

// 	// 	deleteHost(c.PostForm("delete"))
// 	// 	c.Redirect(302, "/registeredhosts")

// 	// }
// 	if actionParam != "start" {
// 		go sendWol(idParam) 
// 		c.Redirect(302, "/registeredhosts")

//		}
//		// if update != "" {
//		// 	currentHost = hosts[update]
//		// 	if currentHost.IsProxmox {
//		// 		currentPVEHost = pveHosts[update]
//		// 	}
//		// 	fmt.Println("CURRENT:", currentHost.Name)
//		// 	c.Redirect(302, "/registeredhosts/edit/"+currentHost.Name)
//		// }
//		// if addhost != "" {
//		// 	c.Redirect(302, "/addhost")
//		// }
//	}
func addHostHandler(c *gin.Context) {
	fmt.Println("adding")
	newhost := createHost(c)
	fmt.Println("username: ", c.PostForm("username"))
	fmt.Println("newhost ", newhost)

	if c.PostForm("username") != "" {
		fmt.Println("adding pve host: ", c.PostForm("username"))
		newPVEhost := models.PVEHost{Name: c.PostForm("name"),
			MacAddress:    newhost.MacAddress,
			IpAddress:     newhost.IpAddress,
			AlternatePort: newhost.AlternatePort,
			OnlineStatus:  newhost.OnlineStatus,
			Credentials: proxmox.Credentials{
				Username: c.PostForm("username"),
				Password: c.PostForm("password"),
			},
		}
		addPVEHost(newPVEhost)
	}
	addHost(newhost)
	c.JSON(http.StatusCreated, newhost)
}

// GetHostByID retrieves a host by its ID from the database
func getHostByIDHandler(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid host ID"})
		return
	}

	var host models.Host
	err = database.Db.QueryRow("SELECT id, name, macaddress, ipaddress, alternateport, onlinestatus, proxmox FROM wolhosts WHERE id=$1", id).
		Scan(&host.Id, &host.Name, &host.MacAddress, &host.IpAddress, &host.AlternatePort, &host.OnlineStatus, &host.IsProxmox)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Host not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": host})
}

// func postEditHostHandler(c *gin.Context) {
// 	mac := strings.ToUpper(c.PostForm("macAddress"))
// 	newMac := strings.ReplaceAll(mac, "-", ":")
// 	ipAdd := net.ParseIP(c.PostForm("ipAddress"))
// 	onlineStatus := pingHost(ipAdd.String())
// 	alternatePort := c.PostForm("alternatePort")
// 	isProxmox := currentHost.IsProxmox
// 	fmt.Println("username: ", c.PostForm("username"))

// 	if c.PostForm("username") != "" {
// 		isProxmox = true

// 		newPVEhost := models.PVEHost{Name: c.PostForm("name"),
// 			MacAddress:    c.PostForm("macAddress"),
// 			IpAddress:     c.PostForm("ipAddress"),
// 			AlternatePort: c.PostForm("alternatePort"),
// 			OnlineStatus:  pingHost((c.PostForm("ipAddress"))),
// 			Credentials: proxmox.Credentials{
// 				Username: c.PostForm("username"),
// 				Password: c.PostForm("password"),
// 			},
// 			ApiCredentials: models.PVEAPICredentials{
// 				Secret:  c.PostForm("secret"),
// 				TokenId: c.PostForm("token"),
// 			},
// 		}
// 		if currentHost.IsProxmox {
// 			updatePVEHost(newPVEhost)
// 		} else {
// 			addPVEHost(newPVEhost)

// 		}

// 	}
// 	updatedHost := models.Host{Name: c.PostForm("name"),
// 		MacAddress:    newMac,
// 		IpAddress:     ipAdd.String(),
// 		AlternatePort: alternatePort,
// 		OnlineStatus:  onlineStatus,
// 		IsProxmox:     isProxmox,
// 	}
// 	updateHost(updatedHost)
// 	c.Redirect(302, "/registeredhosts")
// }

// UpdateHost updates an existing host in the database
func updateHostHandler(c *gin.Context) {
	fmt.Println("update")

	//get Host and params.
	var updateHostParams models.UpdateHostParams
	if err := c.ShouldBindJSON(&updateHostParams); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	host := updateHostParams.Host
	fmt.Println("updatedHost: ", updateHostParams.Host)

	//if perform actions on Host
	if updateHostParams.IsActionUpdate {
		switch updateHostParams.Action {
		// start host
		case models.StartHost:
			if !host.OnlineStatus {
				sendWol(strconv.Itoa(host.Id))
				c.JSON(http.StatusOK, gin.H{"success": true, "message": "Wol Packet Sent"})
				return
			}
		}
	} else {
		//else update Host info in DB
		result, err := database.Db.Exec(
			"UPDATE wolhosts SET name=$1, macaddress=$2, ipaddress=$3, alternateport=$4, onlinestatus=$5, proxmox=$6 WHERE id=$7",
			host.Name, host.MacAddress, host.IpAddress, host.AlternatePort, host.OnlineStatus, host.IsProxmox, host.Id,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
			return
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Host not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Host updated successfully"})
		return
	}

}

// DeleteHost removes a host from the database by its ID
func DeleteHost(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid host ID"})
		return
	}

	wolResult, wolErr := database.Db.Exec("DELETE FROM wolhosts WHERE id=$1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": wolErr.Error()})
		return
	}
	wolRowsAffected, _ := wolResult.RowsAffected()
	if wolRowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "WOL Host not found"})

		return
	}
	if currentHost.IsProxmox {
		pveResult, pveErr := database.Db.Exec("DELETE FROM pvehosts WHERE id=$1", id)
		if pveErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": pveErr.Error()})
			return
		}
		pveRowsAffected, _ := pveResult.RowsAffected()
		if pveRowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "PVE Host not found"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Host deleted successfully"})
}

func addHost(newhost models.Host) {
	rows, rowErr := database.Db.Query(`
	SELECT *
	FROM wolhosts
	WHERE name=$1 OR macaddress=$2 OR ipaddress=$3`, newhost.Name, newhost.MacAddress, newhost.IpAddress)
	fmt.Println("rows:", rows)
	if rowErr != nil {
		panic(rowErr)
	}
	if rows != nil {
		for rows.Next() {
			var host models.Host
			err := rows.Scan(&host.Id, &host.Name, &host.MacAddress, &host.IpAddress, &host.AlternatePort, &host.OnlineStatus, &host.IsProxmox)
			if err != nil {
				panic(err)
			}
			if Compare(host, newhost) {
				break
			}
		}
	}
	if !Contains(wolHosts, newhost) {
		database.Db.Exec(`
		INSERT INTO wolhosts(name, macaddress, ipaddress, alternateport, onlinestatus, proxmox) 
		VALUES($1, $2, $3, $4, $5, $6);`, newhost.Name, newhost.MacAddress, newhost.IpAddress, newhost.AlternatePort, newhost.OnlineStatus, newhost.IsProxmox)
	}
	wolHosts = append(wolHosts, newhost)

}

func getHosts() {
	var currentHosts = []models.Host{}

	rows, rowErr := database.Db.Query(`
	SELECT *
	FROM wolhosts`)
	if rowErr != nil {
		panic(rowErr)
	}
	if rows != nil {
		for rows.Next() {
			var host models.Host
			err := rows.Scan(&host.Id, &host.Name, &host.MacAddress, &host.IpAddress, &host.AlternatePort, &host.OnlineStatus, &host.IsProxmox)
			if err != nil {
				panic(err)
			}
			currentHosts = append(currentHosts, host)
		}
	}
	rows.Close()
	wolHosts = currentHosts
	go checkifHostsOnline()

}

func createHost(c *gin.Context) models.Host {
	mac := strings.ToUpper(c.PostForm("macAddress"))
	newMac := strings.ReplaceAll(mac, "-", ":")
	if _, macErr := net.ParseMAC(newMac); macErr != nil {
		c.Redirect(302, "/registeredhosts")
	}
	ipAdd := net.ParseIP(c.PostForm("ipAddress"))
	onlineStatus := pingHost(ipAdd.String())
	alternatePort := c.PostForm("alternatePort")
	isProxmox := false
	if c.PostForm("username") != "" {
		isProxmox = true
	}
	fmt.Println("isProxmox: ", isProxmox)

	newhost := models.Host{Name: c.PostForm("name"),
		MacAddress:    newMac,
		IpAddress:     ipAdd.String(),
		AlternatePort: alternatePort,
		OnlineStatus:  onlineStatus,
		IsProxmox:     isProxmox,
	}
	return newhost
}
func pingHost(ipAddress string) bool {
	onlineStatus := false
	p := fastping.NewPinger()
	ra, resolveErr := net.ResolveIPAddr("ip4:icmp", ipAddress)
	if resolveErr != nil {
		fmt.Println(resolveErr)
		return onlineStatus
	}
	p.AddIPAddr(ra)
	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		onlineStatus = true
	}
	resolveErr = p.Run()
	if resolveErr != nil {
		fmt.Println(resolveErr)
		return false
	}
	return onlineStatus
}

func sendWol(id string) {
	var host models.Host
	err := database.Db.QueryRow("SELECT id, name, macaddress, ipaddress, alternateport, onlinestatus, proxmox FROM wolhosts WHERE id=$1", id).
		Scan(&host.Id, &host.Name, &host.MacAddress, &host.IpAddress, &host.AlternatePort, &host.OnlineStatus, &host.IsProxmox)
	if err != nil {
		// c.JSON(http.StatusNotFound, gin.H{"error": "Host not found"})
		fmt.Println("host not found")
		return
	}
	if packet, err := wol.New(host.MacAddress); err == nil {
		alternatePort := host.AlternatePort
		if alternatePort != "" {
			sendUDPPacket(packet, host.IpAddress+":"+host.AlternatePort) // specify receiving port
		} else {
			sendUDPPacket(packet, host.IpAddress+":9")

		}

	}
}

func sendUDPPacket(mp *wol.MagicPacket, addr string) (err error) {

	udpAdd, _ := net.ResolveUDPAddr("udp", "255.255.255.255:9")
	bs, err := mp.Marshal()
	var localAddr *net.UDPAddr
	fmt.Println("localAddr:", localAddr)
	fmt.Println("addr:", addr)
	conn, err := net.DialUDP("udp", localAddr, udpAdd)
	if err != nil {
		return err
	}
	defer conn.Close()

	n, err := conn.Write(bs)
	if err == nil && n != 102 {
		err = fmt.Errorf("magic packet sent was %d bytes (expected 102 bytes sent)", n)
	}
	if err != nil {
		return err
	}
	return err
}
func Contains(hosts []models.Host, host models.Host) bool {
	for _, v := range hosts {
		if v.Name == host.Name {
			return true
		}
	}
	return false
}
func Compare(host1 models.Host, host models.Host) bool {

	if host1.Name == host.Name {
		return true
	}
	if host1.MacAddress == host.MacAddress {
		return true
	}
	if host1.IpAddress == host.IpAddress {
		return true
	}
	return false
}
func checkifHostsOnline() {
	fmt.Println("checkOnline:", wolHosts)
	fmt.Println("checkOnline:", pveHosts)

	for _, v := range wolHosts {
		onlineStatus := pingHost(v.IpAddress)
		if onlineStatus != v.OnlineStatus {
			fmt.Println("checkOnlinestat: ", onlineStatus)
			v.OnlineStatus = onlineStatus
			database.Db.Exec(`UPDATE wolhosts
			SET onlineStatus=$1 WHERE name=$2;`, onlineStatus, v.Name)
			if v.IsProxmox {
				fmt.Println("isProxmox")
				database.Db.Exec(`UPDATE pvehosts
				SET onlineStatus=$1 WHERE name=$2;`, onlineStatus, v.Name)
			}
		}

	}
}
