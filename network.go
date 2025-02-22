package usageTracker

import (
	"errors"
	"fmt"
	"github.com/Wifx/gonetworkmanager/v3"
	"github.com/dustin/go-humanize"
	"log"
	"net"
	"strings"
	"time"
)

type (
	NetworkInfo struct {
		Nm                gonetworkmanager.NetworkManager
		WifiInterfaceName string
	}

	Usage struct {
		SSID       string `json:"ssid,omitempty"`
		Download   uint64 `json:"download"`
		Upload     uint64 `json:"upload"`
		TotalUsage uint64 `json:"total_usage"`
		Date       string `json:"date,omitempty"`
	}
	WifiSession struct {
		SSID         string
		LastUpload   uint64
		LastDownload uint64
	}
)

func (u Usage) String() string {
	var usage strings.Builder

	usage.WriteString(fmt.Sprintf("{SSID: %s, Download: %s, Upload: %s, TotalUsage: %s}",
		u.SSID, humanize.Bytes(u.Download), humanize.Bytes(u.Upload), humanize.Bytes(u.TotalUsage)))
	return usage.String()
}

// GetWifiInterfaceName The name of Wi-Fi interface.
func GetWifiInterfaceName() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		for range 5 {
			interfaces, err = net.Interfaces()
			if err == nil {
				break
			}
			time.Sleep(1000 * time.Millisecond)
		}
		if err != nil {
			return "", err
		}
	}
	for _, i := range interfaces {
		if strings.HasPrefix(i.Name, "wl") {
			return i.Name, nil
		}
	}
	return "", errors.New("wifi interface not found")
}

// GetWifiUsage Get the current Wi-Fi usage from the system
func (network *NetworkInfo) GetWifiUsage() (Usage, error) {
	var usage Usage
	wifiDevice, err := GetWifiInterfaceName()
	if err != nil {
		return usage, err
	}
	wifiInterface, err := network.Nm.GetDeviceByIpIface(wifiDevice)
	if err != nil {
		return usage, err
	}

	deviceStatistics, err := gonetworkmanager.NewDeviceStatistics(wifiInterface.GetPath())
	if err != nil {
		return usage, err
	}

	downloadBytes, err := deviceStatistics.GetPropertyRxBytes()
	if err != nil {
		log.Println(err)
		return usage, err
	}

	uploadBytes, err := deviceStatistics.GetPropertyTxBytes()
	if err != nil {
		return usage, err
	}

	wifiName, err := network.GetWifiName()
	if err != nil {
		return usage, err
	}

	usage.SSID = wifiName
	usage.Download = downloadBytes
	usage.Upload = uploadBytes
	usage.TotalUsage = downloadBytes + uploadBytes
	return usage, nil

}

// GetWifiName The name of the connected Wi-Fi (SSID name)
func (network *NetworkInfo) GetWifiName() (string, error) {
	wifiDevice, err := network.Nm.GetDeviceByIpIface(network.WifiInterfaceName)
	if err != nil {
		return "", err
	}
	propertyAvailableConnections, err := wifiDevice.GetPropertyActiveConnection()
	if err != nil {
		return "", err
	}

	ssid, err := propertyAvailableConnections.GetPropertyID()
	if err != nil {
		return "", err
	}
	return ssid, nil
}
