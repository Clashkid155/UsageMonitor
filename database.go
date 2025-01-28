package main

import (
	"fmt"
	"github.com/ostafen/clover/v2"
	"golang.org/x/exp/slices"
	"log"
	"strconv"
	"time"
)

const collectionName = "My Usage"

func (network *NetworkInfo) AddToDB() error {
	/*cDB.DropCollection(collectionName)
	err := cDB.ImportCollection(collectionName, "test.json")
	if err != nil {
		log.Println(err)
	}*/
	now := time.Now()
	today := now.Format("02012006")
	year := strconv.Itoa(now.Year())

	collectionExists, err := cDB.HasCollection(collectionName)
	if err != nil {
		log.Println(err)
	}
	if !collectionExists {
		err := cDB.CreateCollection(collectionName)
		if err != nil {
			return err
		}
	}

	exists, err := cDB.Exists(clover.NewQuery(collectionName).Where(clover.Field("year").Eq(year)))
	if err != nil {
		log.Println(err)
	}
	doc := clover.NewDocument()
	if !exists {
		values := SaveValues{
			Year: year,
			Details: []*Details{
				{SSID: network.usage.SSID,
					Usages: []*saveType{
						{Date: today,
							TotalUsage: network.usage.TotalUsage,
						},
					},
				}},
			Id: clover.NewObjectId(),
		}
		doc.SetAll(values.toMap())
		_, err = cDB.InsertOne(collectionName, doc)
		if err != nil {
			log.Println(err)
		}
		//return nil
	} else {
		updateDbField(fmt.Sprintf("%s", year),
			today, &networkInfo)

	}

	err = cDB.ExportCollection(collectionName, "test.json")
	if err != nil {
		log.Println(err)
	}

	return err
}

func updateDbField(year, date string, net *NetworkInfo) {
	all, err := cDB.FindFirst(clover.NewQuery(collectionName).Where(clover.Field("year").Eq(year)))
	if err != nil {
		log.Println(err)
	}
	var dbResult = &SaveValues{}
	err = all.Unmarshal(dbResult)
	if err != nil {
		log.Println(err)
	}

outer:
	for _, details := range dbResult.Details {
		//TODO: I remember the error. When the day is not today, the else block will execute instead of waiting for the
		// if block to run through all possibility
		/// Check's if the Wi-Fi name already exists. If yes, update its value
		if details.SSID == net.usage.SSID {
			fmt.Println("It worked")
			for index, dbUsage := range details.Usages {
				//fmt.Println("Test stuff:", dbUsage.Date == date, dbUsage.Date, date)
				/// Checks if date in the db is the same as today
				if dbUsage.Date == date {
					var totalUsage = net.usage.TotalUsage
					fmt.Println("Previous usage (DB):", dbUsage.TotalUsage, totalUsage, dbUsage.TotalUsage > totalUsage)
					/// Check if the system usage is higher or lower than the db usage.
					// This should account for a reboot on the same day.
					/// Also checks if the db usage is 0 which means it's a new day, if it's a new day,
					// then we subtract yesterday usage from the system usage.
					/*	if net.usage.TotalUsage < 1000 {

						}*/
					if dbUsage.TotalUsage == 0 && index > 1 {

					} else if totalUsage < dbUsage.TotalUsage { // Should account for a reboot
						totalUsage += dbUsage.TotalUsage
					}

					/*if dbUsage.TotalUsage == 0 && index > 1 {
						fmt.Println(totalUsage-details.Usages[index-1].TotalUsage, details.Usages[index-1].TotalUsage, totalUsage)
						totalUsage = totalUsage - details.Usages[index-1].TotalUsage
					} else if dbUsage.TotalUsage > totalUsage {
						totalUsage = dbUsage.TotalUsage + totalUsage
					}*/
					fmt.Println("Total dbUsage: ", net.usage.TotalUsage, "Same day dbUsage", totalUsage, dbUsage.TotalUsage)
					dbUsage.TotalUsage = 70327992 //totalUsage
					break outer
					/// Checks if date was yesterday or older
				} else if timeCompare(dbUsage.Date, date) && index == (len(details.Usages)-1) { //toInt(dbUsage.Date) < toInt(date) { //Check if the database date is less than today
					fmt.Println("New Day", dbUsage.Date == date, dbUsage.Date, date)
					fmt.Println(net.usage.TotalUsage-dbUsage.TotalUsage, net.usage.TotalUsage)
					details.Usages = append(details.Usages, &saveType{
						Date:       date,
						TotalUsage: 0, //net.usage.TotalUsage - dbUsage.TotalUsage,
					})
					break outer
				}
			}
			fmt.Println("This break")
		}
	}
	/// Checks if the Wi-Fi name is already in the db if no, it adds it.
	if !slices.ContainsFunc(dbResult.Details, func(d *Details) bool {
		return d.SSID == net.usage.SSID
	}) {
		fmt.Println("Else block ran")
		dbResult.Details = append(
			dbResult.Details,
			&Details{
				SSID: net.usage.SSID,
				Usages: []*saveType{
					{
						Date:       date,
						TotalUsage: net.usage.TotalUsage,
					},
				},
			},
		)
	}
	fmt.Println(net.usage.SSID, "Is in dbResult.Details", slices.ContainsFunc(dbResult.Details, func(d *Details) bool {
		return d.SSID == net.usage.SSID
	}))

	dbResult.Id = all.ObjectId()
	fmt.Println("Update:", dbResult, len(dbResult.Details[0].Usages))
	err = cDB.UpdateById(collectionName, dbResult.Id, dbResult.toMap())
	if err != nil {
		log.Println(err)
	}
	fmt.Println("Update ran", year, date)
}

// Checks if todayDate was yesterday or older when compared
// against dbDate.
func timeCompare(dbDate, todayDate string) bool {
	dbTime, err := time.Parse("02012006", dbDate)
	if err != nil {
		log.Println(err)
	}
	todayTime, err := time.Parse("02012006", todayDate)
	if err != nil {
		log.Println(err)
	}
	return dbTime.Before(todayTime)
}
