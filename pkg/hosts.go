package router

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/luthermonson/go-proxmox"
	"github.com/sabhiram/go-wol/wol"
	"github.com/tatsushid/go-fastping"
	database "github.com/tophmayor/wakehost/db"
	models "github.com/tophmayor/wakehost/models"
)

func checkifHostsOnline() {
	for _, v := range hosts {
		onlineStatus := pingHost(v.IpAddress)
		if onlineStatus != v.OnlineStatus {
			v.OnlineStatus = onlineStatus
		}
		database.Db.Exec(`UPDATE wolhosts
			SET onlineStatus=$1 WHERE name=$2;`, onlineStatus, v.Name)
	}
}
func getAddHostHandler(c *gin.Context) {
	if database.ConfigNeeded {
		c.Redirect(302, "/setup")
	} else {
		c.HTML(http.StatusOK, "addhosts.html", gin.H{"Hosts": hosts})
	}
}

func getEditHostHandler(c *gin.Context) {
	if database.ConfigNeeded {
		c.Redirect(302, "/setup")
	} else {
		c.HTML(http.StatusOK, "edithosts.html", gin.H{"Host": currentHost, "PVEHost": currentPVEHost})
	}
}

func getHostsHandler(c *gin.Context) {
	if database.ConfigNeeded {
		c.Redirect(302, "/setup")
	} else {
		getHosts()
		currentTime := time.Now().Format(time.RFC850)
		c.HTML(http.StatusOK, "hosts.html", gin.H{"Hosts": hosts, "CurrentTime": currentTime})
		fmt.Println("hosts:", hosts)
	}
}

func postHostHandler(c *gin.Context) {
	wol := c.PostForm("wol")
	del := c.PostForm("delete")
	update := c.PostForm("update")
	addhost := c.PostForm("addhost")

	if del != "" {
		fmt.Println("del:", del)

		deleteHost(c.PostForm("delete"))
		c.Redirect(302, "/registeredhosts")

	}
	if wol != "" {
		sendWol(c.PostForm("wol"))
		c.Redirect(302, "/registeredhosts")

	}
	if update != "" {
		currentHost = hosts[update]
		if currentHost.IsProxmox {
			currentPVEHost = pveHosts[update]
		}
		fmt.Println("CURRENT:", currentHost.Name)
		c.Redirect(302, "/registeredhosts/edit/"+currentHost.Name)
	}
	if addhost != "" {
		c.Redirect(302, "/addhost")
	}
}
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
	c.Redirect(302, "/registeredhosts")
}
func postEditHostHandler(c *gin.Context) {
	mac := strings.ToUpper(c.PostForm("macAddress"))
	newMac := strings.ReplaceAll(mac, "-", ":")
	ipAdd := net.ParseIP(c.PostForm("ipAddress"))
	onlineStatus := pingHost(ipAdd.String())
	alternatePort := c.PostForm("alternatePort")
	isProxmox := currentHost.IsProxmox
	fmt.Println("username: ", c.PostForm("username"))

	if c.PostForm("username") != "" {
		isProxmox = true

		newPVEhost := models.PVEHost{Name: c.PostForm("name"),
			MacAddress:    c.PostForm("macAddress"),
			IpAddress:     c.PostForm("ipAddress"),
			AlternatePort: c.PostForm("alternatePort"),
			OnlineStatus:  pingHost((c.PostForm("ipAddress"))),
			Credentials: proxmox.Credentials{
				Username: c.PostForm("username"),
				Password: c.PostForm("password"),
			},
			ApiCredentials: models.PVEAPICredentials{
				Secret:  c.PostForm("secret"),
				TokenId: c.PostForm("token"),
			},
		}
		if currentHost.IsProxmox {
			updatePVEHost(newPVEhost)
		} else {
			addPVEHost(newPVEhost)

		}

	}
	updatedHost := models.Host{Name: c.PostForm("name"),
		MacAddress:    newMac,
		IpAddress:     ipAdd.String(),
		AlternatePort: alternatePort,
		OnlineStatus:  onlineStatus,
		IsProxmox:     isProxmox,
	}
	updateHost(updatedHost)
	c.Redirect(302, "/registeredhosts")
}
func deleteHost(hostName string) {
	fmt.Println("deleting?", hostName)
	_, errDel := database.Db.Exec(`DELETE FROM wolhosts
	 WHERE name=$1;`, hostName)
	if errDel != nil {
		fmt.Println("panic wol del")
		// panic(errDel)
	}
	if currentHost.IsProxmox {
		_, errDel := database.Db.Exec(`DELETE FROM pvehosts
	 WHERE name=$1;`, hostName)
		if errDel != nil {
			fmt.Println("panic pve del")
			// panic(errDel)
		}
	}
}

func updateHost(host models.Host) {
	columns, values := findColumnDiffs(currentHost, host)

	cmd := ""
	for i, v := range columns {
		if v == "macAddress" || v == "ipAddress" {
			if v == "macAddress" {
				cmd += v + "=" + `CAST('` + values[i] + `' AS macaddr)`

			} else {
				cmd += v + "=" + `CAST('` + values[i] + `' AS inet)`
			}
		} else {
			cmd += v + `='` + values[i] + `'`
		}
		if i < len(columns)-1 {
			cmd += ", "
		}
	}
	var final = `UPDATE wolhosts
	SET ` + cmd + ` WHERE name='` + currentHost.Name + `';`
	_, err := database.Db.Exec(final)
	if err != nil {
		fmt.Println("panic")

		panic(err)
	}
}

func findColumnDiffs(oldData models.Host, newData models.Host) ([]string, []string) {
	var columns []string
	var values []string

	if oldData.Name != newData.Name {
		columns = append(columns, "name")
		values = append(values, newData.Name)
	}
	if oldData.IpAddress != newData.IpAddress {
		columns = append(columns, "ipAddress")
		values = append(values, newData.IpAddress)
	}
	if oldData.MacAddress != newData.MacAddress {
		columns = append(columns, "macAddress")
		values = append(values, newData.MacAddress)
	}
	if oldData.AlternatePort != newData.AlternatePort {
		columns = append(columns, "alternatePort")
		alternatePort := newData.AlternatePort
		values = append(values, alternatePort)
	}
	if oldData.IsProxmox != newData.IsProxmox {
		columns = append(columns, "proxmox")
		isProxmox := newData.IsProxmox
		if isProxmox {
			values = append(values, `true`)
		} else {
			values = append(values, `false`)
		}
	}
	return columns, values
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
			err := rows.Scan(&host.HostId, &host.Name, &host.MacAddress, &host.IpAddress, &host.AlternatePort, &host.OnlineStatus, &host.IsProxmox)
			if err != nil {
				panic(err)
			}
			if Compare(host, newhost) {
				break
			}
		}
	}
	if !Contains(hosts, newhost) {
		database.Db.Exec(`
		INSERT INTO wolhosts(name, macaddress, ipaddress, alternateport, onlinestatus, proxmox) 
		VALUES($1, $2, $3, $4, $5, $6);`, newhost.Name, newhost.MacAddress, newhost.IpAddress, newhost.AlternatePort, newhost.OnlineStatus, newhost.IsProxmox)
	}
	hosts[newhost.Name] = newhost

}

func getHosts() {
	var currentHosts = map[string]models.Host{}

	rows, rowErr := database.Db.Query(`
	SELECT *
	FROM wolhosts`)
	if rowErr != nil {
		panic(rowErr)
	}
	if rows != nil {
		for rows.Next() {
			var host models.Host
			err := rows.Scan(&host.HostId, &host.Name, &host.MacAddress, &host.IpAddress, &host.AlternatePort, &host.OnlineStatus, &host.IsProxmox)
			if err != nil {
				panic(err)
			}
			currentHosts[host.Name] = host
		}
	}
	rows.Close()
	hosts = currentHosts
	checkifHostsOnline()

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
func sendWol(key string) {
	host := hosts[key]
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
func Contains(hosts map[string]models.Host, host models.Host) bool {
	for k, v := range hosts {
		if k == host.Name && v == host {
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
