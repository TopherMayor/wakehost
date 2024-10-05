package router

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/luthermonson/go-proxmox"
	"github.com/sabhiram/go-wol/wol"
	"github.com/tatsushid/go-fastping"
)

// // func ListFiles(ctx *gin.Context) {
// // 	path := getDir()
// // 	entries, err := os   .ReadDir(path)
// // 	if err != nil {
// // 		log.Fatal(err)
// // 	}
// // 	output := []map[string]string{}
// // 	for _, e := range entries {
// // 		info, _ := e.Info()
// // 		var size string
// // 		if info.Size() < 1024 {
// // 			size = strconv.Itoa(int(info.Size())) + " bytes"
// // 		} else if 1024 < info.Size() && info.Size() < 1048576 {
// // 			size = strconv.FormatFloat(float64(info.Size())/float64(1024), 'f', 2, 64) + " Kbs"
// // 		} else {
// // 			size = strconv.FormatFloat(float64(info.Size())/float64(1048576), 'f', 2, 64) + " Mbs"
// // 		}
// // 		output = append(output, map[string]string{
// // 			"name": e.Name(),
// // 			"size": size,
// // 		})

// // 	}
// // 	ctx.HTML(http.StatusOK, "files.html", gin.H{"files": output})
// // }

// // func UploadFiles(ctx *gin.Context) {
// // 	form, _ := ctx.MultipartForm()
// // 	files := form.File["file"]
// // 	path := getDir()
// // 	for _, file := range files {
// // 		err := ctx.SaveUploadedFile(file, filepath.Join(path, file.Filename))
// // 		if err != nil {
// // 			displayError(ctx, "Error in Uploading Multiple file ,as the storage dir. not found ", err)
// // 		}
// // 	}

// // 	fmt.Println(ctx.PostForm("key"))
// // 	// ctx.Redirect(http.StatusFound, "/files")
// // 	ctx.JSON(http.StatusOK, map[string]string{"status": "Accepted"})

// // }

// // func DownloadFile(ctx *gin.Context) {
// // 	path, err := filepath.Abs("sync.io-cache")
// // 	if err != nil {
// // 		log.Fatal(err)
// // 	}
// // 	fileName := ctx.Param("filename")

// // 	ctx.FileAttachment(filepath.Join(path, fileName), fileName)
// // }
// // func DownloadAllFiles(ctx *gin.Context) {
// // 	path, _ := filepath.Abs("sync.io-cache")
// // 	files, err := getFilesInDir(path)
// // 	if err != nil {
// // 		ctx.String(http.StatusInternalServerError, "Failed to read directory: %v", err)
// // 		return
// // 	}

// // 	zipFile := "files.zip"
// // 	if err := zipFiles(zipFile, files); err != nil {
// // 		ctx.String(http.StatusInternalServerError, "Failed to zip files: %v", err)
// // 		return
// // 	}

// // 	ctx.FileAttachment(zipFile, "files.zip")
// // 	defer os.Remove(zipFile)

// // }
// // func PreviewFile(ctx *gin.Context) {
// // 	path, err := filepath.Abs("sync.io-cache")
// // 	if err != nil {
// // 		log.Fatal(err)
// // 	}
// // 	fileName := ctx.Param("filename")

// //		ctx.File(filepath.Join(path, fileName))
// //	}
// //
// //	func DeleteAllFiles(ctx *gin.Context) {
// //		path, err := filepath.Abs("sync.io-cache")
// //		if err != nil {
// //			log.Fatal(err)
// //		}
// //		entries, _ := os.ReadDir(path)
// //		for _, e := range entries {
// //			os.Remove(filepath.Join(path, e.Name()))
// //		}
// //		ctx.Redirect(http.StatusFound, "/files")
// //	}
// //
// //	func DeleteFile(ctx *gin.Context) {
// //		path, err := filepath.Abs("sync.io-cache")
// //		if err != nil {
// //			log.Fatal(err)
// //		}
// //		fileName := ctx.Param("filename")
// //		e := os.Remove(filepath.Join(path, fileName))
// //		if e == nil {
// //			ctx.Redirect(http.StatusFound, "/files")
// //		} else {
// //			panic(e)
// //		}
// //	}

// //	type WakeOnLanHost struct {
// //		MacAddress string     `json:"MacAddress"`
// //		Name       string     `json:"Name"`
// //		LatestUsed *time.Time `json:"LatestUsed"`
// //		OnlineStatus bool	  `json:"OnlineStatus"`
// //	}
// //
// //	func addHostHandler(host WakeOnLanHost, hostStore *HostStore) {
// //		addHost(host, hostStore)
// //	}
func loadHosts() (*HostStore, error) {
	content, err := os.ReadFile("hosts.json")
	if err != nil {
		log.Fatal(err)
	}
	var hostStore HostStore
	err = json.Unmarshal(content, &hostStore)
	fmt.Println(hostStore.Hosts["pve"].IpAddress)
	return &hostStore, err

}
func loadPVEHosts() (*PVEHostStore, error) {
	content, err := os.ReadFile("pvehosts.json")
	if err != nil {
		log.Fatal(err)
	}
	var hostStore PVEHostStore
	err = json.Unmarshal(content, &hostStore)
	// fmt.Println(hostStore.Hosts["pve"].IpAddress)
	return &hostStore, err

}

func addHost(host WakeOnLanHost, hostStore *HostStore) {
	hostStore.Hosts[host.Name] = host
	content, _ := json.Marshal(hostStore)
	err := os.WriteFile("hosts.json", content, 0644)
	if err != nil {
		log.Fatal(err)
	}

}

func addPVEHost(host PVEHost, hostStore *PVEHostStore) {
	hostStore.Hosts[host.Name] = host
	content, _ := json.Marshal(hostStore)
	err := os.WriteFile("pvehosts.json", content, 0644)
	if err != nil {
		log.Fatal(err)
	}

}
func checkifHostsOnline(hostStore *HostStore) {
	for k, v := range hostStore.Hosts {
		onlineStatus := pingHost(v.IpAddress)
		hostStore.Hosts[k] = WakeOnLanHost{Name: v.Name,
			MacAddress:   v.MacAddress,
			IpAddress:    v.IpAddress,
			OnlineStatus: onlineStatus,
		}
	}
	content, _ := json.Marshal(hostStore)
	err := os.WriteFile("hosts.json", content, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func pingHost(ipAddress string) bool {
	onlineStatus := false
	p := fastping.NewPinger()
	ra, err := net.ResolveIPAddr("ip4:icmp", ipAddress)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	p.AddIPAddr(ra)
	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		// fmt.Printf("IP Addr: %s receive, RTT: %v\n", addr.String(), rtt)
		onlineStatus = true
	}
	// p.OnIdle = func() {
	// fmt.Println("finish")
	// }
	err = p.Run()
	if err != nil {
		fmt.Println(err)
	}
	return onlineStatus
}

func sendUDPPacket(mp *wol.MagicPacket, addr string) (err error) {
	udpAdd, _ := net.ResolveUDPAddr("udp", "255.255.255.255:9")
	bs, err := mp.Marshal()
	fmt.Printf("ip: %s\n", addr)
	fmt.Printf("udp: %s\n", udpAdd)
	var localAddr *net.UDPAddr
	fmt.Printf("local: %s\n", localAddr)

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
func sendWol(hostStore *HostStore, key string) {
	fmt.Printf("name: %s\n", key)
	host := hostStore.Hosts[key]
	if packet, err := wol.New(host.MacAddress); err == nil {
		if host.AlternatePort != "" {
			// sendUDPPacket(packet, host.IpAddress+":"+host.AlternatePort) // specify receiving port
		} else {
			sendUDPPacket(packet, host.IpAddress+":9")
			fmt.Printf("ip: %s\n", host.IpAddress)
			fmt.Printf("mac: %s\n", host.MacAddress)

		}

	}
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
