package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/Wifx/gonetworkmanager/v3"
	usageTracker "github.com/clashkid155/usage-monitor"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

var (
	sqlDb             *sql.DB
	deviceNetworkInfo usageTracker.NetworkInfo
	currentSession    *usageTracker.WifiSession
	configPath        string
)

func init() {
	var err error
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	configPath, err = os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}
	configPath = filepath.Join(configPath, "Usage Monitor")
	err = os.MkdirAll(configPath, 0750)
	if err != nil {
		log.Fatalln(fmt.Errorf("unable to create usage monitor directory: %s", err))
	}

	deviceNetworkInfo.Nm, err = gonetworkmanager.NewNetworkManager()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("db location:", fmt.Sprintf("file:%s",
		filepath.Join(configPath, "db.sqlite")))
	sqlDb, err = sql.Open("sqlite3", fmt.Sprintf("file:%s",
		filepath.Join(configPath, "db.sqlite")))
	if err != nil {
		log.Fatal(err)
	}
	err = usageTracker.CreateTable(sqlDb)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	var err error
	deviceNetworkInfo.WifiInterfaceName, err = usageTracker.GetWifiInterfaceName()
	if err != nil {
		log.Println(err)
		return
	}
	wifiName, err := deviceNetworkInfo.GetWifiName()
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println("CONNECTED TO:", wifiName)
	wifiUsage, err := deviceNetworkInfo.GetWifiUsage()
	if err != nil {
		log.Println(err)
		return
	}
	currentSession = &usageTracker.WifiSession{
		SSID:         wifiUsage.SSID,
		LastUpload:   wifiUsage.Upload,
		LastDownload: wifiUsage.Download,
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	go httpListener()
	RunUntilInterrupt(&deviceNetworkInfo, time.NewTicker(10*time.Second), ch)

}

// SetUsageInDb Should run when connected to a new Wi-Fi
//
// This function should handle adding new Wi-Fi to the database
func SetUsageInDb(usage *usageTracker.Usage) {
	_, err := usageTracker.GetUsageBySsid(sqlDb, usage)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = usageTracker.InsertUsage(sqlDb, &usageTracker.Usage{
				SSID:     usage.SSID,
				Download: 0,
				Upload:   0,
			})
			if err != nil {
				log.Println(err)
				return
			}
			currentSession.SSID = usage.SSID
		}
		return
	}
	currentSession = &usageTracker.WifiSession{
		SSID:         usage.SSID,
		LastUpload:   usage.Upload,
		LastDownload: usage.Download,
	}
}

// RefreshDbUsage Should run periodically to query the
// system usage and set in the database.
func RefreshDbUsage(network *usageTracker.NetworkInfo) {
	// System usage
	usage, err := network.GetWifiUsage()
	if err != nil {
		log.Println(err)
		return
	}
	// Usage based on current ssid from the db
	usageBySsid, err := usageTracker.GetUsageBySsid(sqlDb, &usage)

	if errors.Is(err, sql.ErrNoRows) {
		err = usageTracker.InsertUsage(sqlDb, &usageTracker.Usage{
			SSID: usage.SSID,
		})
		if err != nil {
			log.Println("RefreshDbUsage Insert Usage Error:", err)
			return
		}
		return
	} else if err != nil {
		log.Println(err)
		return
	}

	downloadUsage := usage.Download - currentSession.LastDownload + usageBySsid.Download
	uploadUsage := usage.Upload - currentSession.LastUpload + usageBySsid.Upload

	if usageBySsid.SSID == usage.SSID {
		err = usageTracker.UpdateUsage(sqlDb, &usageTracker.Usage{
			SSID:     usage.SSID,
			Download: downloadUsage,
			Upload:   uploadUsage,
		})
		if err != nil {
			log.Println(err)
			return
		}
	}

	currentSession.LastDownload = usage.Download
	currentSession.LastUpload = usage.Upload
}

// RunUntilInterrupt Keep running until an error occurs or a system interrupt is sent.
func RunUntilInterrupt(network *usageTracker.NetworkInfo, ticker *time.Ticker, exit chan os.Signal) {
	for {
		select {
		case <-exit:
			network.Nm.Unsubscribe()
			ticker.Stop()
			close(exit)
			fmt.Println("\nExiting Usage Monitor...")
			return

		case <-ticker.C:
			nmState, err := network.Nm.State()
			if err != nil {
				log.Println(err)
			}

			if nmState == gonetworkmanager.NmStateConnectedGlobal {
				RefreshDbUsage(network)
			}
		case sub := <-network.Nm.Subscribe():
			//  https://networkmanager.dev/docs/api/latest/spec.html Provided all the help I needed
			//  By monitoring the Device state change, we can get the various states the interface has taken.
			//  Details on sub.Body https://networkmanager.dev/docs/api/latest/gdbus-org.freedesktop.NetworkManager.Device.html

			if sub.Name == fmt.Sprint(gonetworkmanager.DevicePropertyState, "Changed") && sub.Body[0] == uint32(100) {

				wifiName, err := network.GetWifiName()
				if err != nil {
					log.Println(err)
				}

				fmt.Println("\nSWITCHED TO:", wifiName)

				wifiUsage, err := network.GetWifiUsage()
				if err != nil {
					log.Println(err)
				}
				SetUsageInDb(&wifiUsage)

			}
		}

	}
}
