package main

import (
	usageTracker "UsageMonitor"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Wifx/gonetworkmanager/v3"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var sqlDb *sql.DB
var deviceNetworkInfo usageTracker.NetworkInfo

func init() {
	var err error
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	deviceNetworkInfo.Nm, err = gonetworkmanager.NewNetworkManager()
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}
	sqlDb, err = sql.Open("sqlite3", "./wifi_usage_tracker.sqlite")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	err = usageTracker.CreateTable(sqlDb)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func main() {
	var err error
	deviceNetworkInfo.WifiInterfaceName, err = deviceNetworkInfo.GetWifiDevice()
	if err != nil {
		return
	}
	wifiName, err := deviceNetworkInfo.GetWifiName()
	if err != nil {
		return
	}
	fmt.Println("CONNECTED TO:", wifiName)

	/*	ssid, err := usageTracker.GetUsageBySsid(sqlDb, &deviceNetworkInfo.Usage)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				log.Println(err)
				// return
			}
			log.Println(err)
		}
		fmt.Println("Exist sql db:", ssid, "Error:", err) */
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	RunUntilInterrupt(&deviceNetworkInfo, time.NewTicker(1*time.Minute), ch)

}

// SetUsageInDb Should run when connected to a new Wi-Fi
//
// This function should handle adding new Wi-Fi to the database
func SetUsageInDb(usage *usageTracker.Usage) {
	_, err := usageTracker.GetUsageBySsid(sqlDb, usage)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = usageTracker.InsertUsage(sqlDb, usage)
			if err != nil {
				log.Println(err)
				return
			}
		}
		return
	}
}

// RefreshDbUsage Should run periodically to query the
// system usage and set in the database.
func RefreshDbUsage(network *usageTracker.NetworkInfo) {
	usage, err := network.GetWifiUsage()
	if err != nil {
		log.Println(err)
		return
	}

	usageBySsid, err := usageTracker.GetUsageBySsid(sqlDb, &usage)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Println(err)
	}
	if errors.Is(err, sql.ErrNoRows) {
		err = usageTracker.InsertUsage(sqlDb, &usage)
		if err != nil {
			log.Println(err)
			return
		}
	}

	if usageBySsid.SSID == usage.SSID {
		/*usageBySsid.Upload += usage.Upload
		usageBySsid.Download += usage.Download*/
		err = usageTracker.UpdateUsage(sqlDb, &usage)
		if err != nil {
			log.Println(err)
			return
		}
	}
}

// RunUntilInterrupt Keep running until an error occurs or a system interrupt is sent.
func RunUntilInterrupt(network *usageTracker.NetworkInfo, ticker *time.Ticker, exit chan os.Signal) {
	for {
		select {
		case <-exit:
			fmt.Println("\nKilling WiFi Usage Monitor...")
			network.Nm.Unsubscribe()
			ticker.Stop()
			close(exit)
			return

		case <-ticker.C:
			nmState, err := network.Nm.State()
			if err != nil {
				log.Println(err)
			}

			if nmState == gonetworkmanager.NmStateConnectedGlobal {
				fmt.Println("It ran after 10 minutes.")
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
