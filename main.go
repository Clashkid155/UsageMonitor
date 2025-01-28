package main

import (
	"encoding/json"
	"fmt"
	"github.com/Wifx/gonetworkmanager/v2"
	"github.com/dustin/go-humanize"
	c "github.com/ostafen/clover/v2"
	"github.com/shirou/gopsutil/v3/net"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

type (
	NetworkInfo struct {
		nm                gonetworkmanager.NetworkManager
		usage             Usage
		wifiInterfaceName string
	}
	Usage struct {
		SSID       string
		Download   string //uint64
		Upload     string //uint64
		TotalUsage uint64
	}
)

var networkInfo NetworkInfo
var cDB *c.DB

func init() {
	var err error
	networkInfo.nm, err = gonetworkmanager.NewNetworkManager()
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}

	cDB, err = c.Open("clover-last")
	if err != nil {
		log.Println(err)
	}

}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	/*
		var current uint64
		c := time.Tick(1 * time.Second)
		for now := range c {
			s, _ := net.IOCounters(false)
			bytesSent := s[0].BytesSent
			fmt.Println(now, bytesSent-current)
			current = bytesSent
		}*/

	connections, err := net.IOCounters(true)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(connections[2])
loop:
	for _, stat := range connections {
		if strings.HasPrefix(stat.Name, "wl") {
			networkInfo.wifiInterfaceName = stat.Name
			networkInfo.usage.Download = humanize.Bytes(stat.BytesRecv)
			networkInfo.usage.Upload = humanize.Bytes(stat.BytesSent)
			networkInfo.usage.TotalUsage = stat.BytesSent + stat.BytesRecv
			break loop
		}
	}

	networkInfo.getWifiName()
	state, err := networkInfo.nm.State()
	if err != nil {
		log.Println(err)
	}
	fmt.Println(networkInfo.usage, state)

	/*nm.GetPropertyActivatingConnection() not working

	I have two ways to get the Wi-Fi ssid, the first methods require I know the Wi-Fi interface name,
	while the second method is to get all the physical device then loop through them and call the

	*/

	defer cDB.Close()
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	err = networkInfo.AddToDB()
	if err != nil {
		log.Println(err)
	}
	go httpListener()
	for {
		select {
		case <-ch:
			networkInfo.nm.Unsubscribe()
			fmt.Println("\nKilling WiFi Usage Tracker")
			return
		case sub := <-networkInfo.nm.Subscribe():
			/// https://networkmanager.dev/docs/api/latest/spec.html Provided all the help I needed
			/// By monitoring the Device state change, we can get the various states the interface has taken.
			/// Details on sub.Body https://networkmanager.dev/docs/api/latest/gdbus-org.freedesktop.NetworkManager.Device.html
			//fmt.Println(sub.Name)
			if sub.Name == fmt.Sprint(gonetworkmanager.DevicePropertyState, "Changed") && sub.Body[0] == uint32(100) {
				//fmt.Printf("Body is %T %T\n", sub.Body, sub.Body[0])
				fmt.Println("New SSID:", networkInfo.getWifiName().usage.SSID)
				err := networkInfo.AddToDB()
				if err != nil {
					log.Println(err)
				}

			}
			//time.Sleep(time.Second)

		}

	}

}

func (network *NetworkInfo) getWifiName() *NetworkInfo {
	wifiDevice, err := network.nm.GetDeviceByIpIface(network.wifiInterfaceName)
	if err != nil {
		log.Println(err)
	}
	propertyAvailableConnections, err := wifiDevice.GetPropertyActiveConnection()
	if err != nil {
		log.Println(err)
	}

	network.usage.SSID, err = propertyAvailableConnections.GetPropertyID()
	if err != nil {
		log.Println(err)
	}

	return network
}

func (u Usage) String() string {
	marshal, err := json.Marshal(u)
	if err != nil {
		log.Println(err)
	}
	return string(marshal)
}
