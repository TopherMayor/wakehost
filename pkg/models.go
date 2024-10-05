// package router

// type StringBool struct {
// 	Str  string
// 	Flag bool
// }

// type Channel struct {
// 	password          string
// 	connected_devices map[string]map[string]StringBool
// }

//	func setPassword(password string) {
//		if password != "" {
//			channel.password = hashAndSalt([]byte(password))
//		}
//	}
//
//	func verifyPassword(password string) bool {
//		hash := hashAndSalt([]byte(password))
//		if hash == channel.password {
//			return false
//		} else {
//			return true
//		}
//	}
package router

import (
	"github.com/luthermonson/go-proxmox"
)

type WakeOnLanHost struct {
	Name          string `json:"name"`
	MacAddress    string `json:"macAddress"`
	IpAddress     string `json:"ipAddress"`
	AlternatePort string `json:"alternatePort"`
	OnlineStatus  bool   `json:"onlineStatus"`
}

type PVEHost struct {
	Name string `json:"name"`
	// MacAddress    string `json:"macAddress"`
	IpAddress   string `json:"ipAddress"`
	Port        string `json:"Port"`
	Credentials proxmox.Credentials
	// AlternatePort string `json:"alternatePort"`
	// OnlineStatus  bool   `json:"onlineStatus"`
}

type HostStore struct {
	Hosts map[string]WakeOnLanHost
}
type PVEHostStore struct {
	Hosts map[string]PVEHost
}

// func loadHosts() (*HostStore, error) {
// 	content, err := os.ReadFile("hosts.json")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	var hostStore HostStore
// 	err = json.Unmarshal(content, &hostStore)
// 	fmt.Println(hostStore.Hosts["pve"].IpAddress)
// 	return &hostStore, err

// }
// func loadPVEHosts() (*PVEHostStore, error) {
// 	content, err := os.ReadFile("pvehosts.json")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	var hostStore PVEHostStore
// 	err = json.Unmarshal(content, &hostStore)
// 	// fmt.Println(hostStore.Hosts["pve"].IpAddress)
// 	return &hostStore, err

// }

// func addHost(host WakeOnLanHost, hostStore *HostStore) {
// 	hostStore.Hosts[host.Name] = host
// 	content, _ := json.Marshal(hostStore)
// 	err := os.WriteFile("hosts.json", content, 0644)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// }

// func addPVEHost(host PVEHost, hostStore *PVEHostStore) {
// 	hostStore.Hosts[host.Name] = host
// 	content, _ := json.Marshal(hostStore)
// 	err := os.WriteFile("pvehosts.json", content, 0644)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// }
// func checkifHostsOnline(hostStore *HostStore) {
// 	for k, v := range hostStore.Hosts {
// 		onlineStatus := pingHost(v.IpAddress)
// 		hostStore.Hosts[k] = WakeOnLanHost{Name: v.Name,
// 			MacAddress:   v.MacAddress,
// 			IpAddress:    v.IpAddress,
// 			OnlineStatus: onlineStatus,
// 		}
// 	}
// 	content, _ := json.Marshal(hostStore)
// 	err := os.WriteFile("hosts.json", content, 0644)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }

// func pingHost(ipAddress string) bool {
// 	onlineStatus := false
// 	p := fastping.NewPinger()
// 	ra, err := net.ResolveIPAddr("ip4:icmp", ipAddress)
// 	if err != nil {
// 		fmt.Println(err)
// 		os.Exit(1)
// 	}
// 	p.AddIPAddr(ra)
// 	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
// 		// fmt.Printf("IP Addr: %s receive, RTT: %v\n", addr.String(), rtt)
// 		onlineStatus = true
// 	}
// 	// p.OnIdle = func() {
// 	// fmt.Println("finish")
// 	// }
// 	err = p.Run()
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	return onlineStatus
// }

// func sendUDPPacket(mp *wol.MagicPacket, addr string) (err error) {
// 	udpAdd, _ := net.ResolveUDPAddr("udp", "255.255.255.255:9")
// 	bs, err := mp.Marshal()
// 	fmt.Printf("ip: %s\n", addr)
// 	fmt.Printf("udp: %s\n", udpAdd)
// 	var localAddr *net.UDPAddr
// 	fmt.Printf("local: %s\n", localAddr)

// 	conn, err := net.DialUDP("udp", localAddr, udpAdd)
// 	if err != nil {
// 		return err
// 	}
// 	defer conn.Close()

// 	n, err := conn.Write(bs)
// 	if err == nil && n != 102 {
// 		err = fmt.Errorf("magic packet sent was %d bytes (expected 102 bytes sent)", n)
// 	}
// 	if err != nil {
// 		return err
// 	}
// 	return err
// }
// func sendWol(hostStore *HostStore, key string) {
// 	fmt.Printf("name: %s\n", key)
// 	host := hostStore.Hosts[key]
// 	if packet, err := wol.New(host.MacAddress); err == nil {
// 		if host.AlternatePort != "" {
// 			// sendUDPPacket(packet, host.IpAddress+":"+host.AlternatePort) // specify receiving port
// 		} else {
// 			sendUDPPacket(packet, host.IpAddress+":9")
// 			fmt.Printf("ip: %s\n", host.IpAddress)
// 			fmt.Printf("mac: %s\n", host.MacAddress)

// 		}

// 	}
// }
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
// func useVM(node *proxmox.Node, key string, start bool) {
// 	if start {
// 		startVM(node, key)
// 	} else {
// 		stopVM(node, key)
// 	}
// }

// func (host WakeOnLanHost) ResolveResourceName() (string, error) {
// 	resourceName := strings.Trim(host.MacAddress, " \t\r\n\000")
// 	if len(resourceName) == 0 {
// 		return "", errors.New("mac Address is Empty")
// 	}
// 	resourceName = strings.ToUpper(resourceName)
// 	resourceName = strings.ReplaceAll(resourceName, ":", "-")

// 	if _, err := net.ParseMAC(resourceName); err != nil {
// 		return "", errors.New("invalid mac address")
// 	}

// 	return resourceName, nil
// }

// const WakeOnLanHostCollectionName = "wake_on_lan_hosts"
