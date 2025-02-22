package usageTracker

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

// CreateTable Create the database table if it doesn't exist.
func CreateTable(db *sql.DB) error {
	const tableQuery = `create table if not exists wifi_usage
(
    id INTEGER primary key autoincrement,
    date TEXT not null,
    ssid TEXT not null,
    upload_usage INTEGER not null,
    download_usage INTEGER not null
);`
	_, err := db.Exec(tableQuery)
	if err != nil {
		return err
	}
	return nil
}

// InsertUsage Add a new Wi-Fi usage to the database.
func InsertUsage(db *sql.DB, usage *Usage) error {
	insertQuery := `insert into wifi_usage (date, ssid, upload_usage, download_usage) values (?, ?, ?, ?)`

	_, err := db.Exec(insertQuery, getFormattedTime(), usage.SSID, usage.Upload, usage.Download)
	if err != nil {
		return fmt.Errorf("insert Usage: %v", err)
	}
	return nil
}

// UpdateUsage Update the database based on ssid and date.
func UpdateUsage(db *sql.DB, usage *Usage) error {
	updateQuery := `update wifi_usage set upload_usage = ?, download_usage = ? where  date = ? and ssid = ?;`

	_, err := db.Exec(updateQuery, usage.Upload, usage.Download, getFormattedTime(), usage.SSID)
	if err != nil {
		return fmt.Errorf("insert Usage: %v", err)
	}
	return nil
}

// DeleteUsage Delete a usage based on ssid and date.
func DeleteUsage(db *sql.DB, usage *Usage) error {
	deleteQuery := `delete from wifi_usage where  date = ? and ssid = ?;`

	_, err := db.Exec(deleteQuery, getFormattedTime(), usage.SSID)
	if err != nil {
		return fmt.Errorf("insert Usage: %v", err)
	}
	return nil
}

// GetUsageBySsid Get usage based on the date and ssid
func GetUsageBySsid(db *sql.DB, usage *Usage) (Usage, error) {
	usageQuery := `select date, ssid, upload_usage, download_usage from wifi_usage where date = ? and ssid = ?;`
	var dbUsage Usage
	row := db.QueryRow(usageQuery, getFormattedTime(), usage.SSID)
	err := row.Scan(&dbUsage.Date, &dbUsage.SSID, &dbUsage.Upload, &dbUsage.Download)
	if err != nil {
		return dbUsage, err
	}
	dbUsage.TotalUsage = dbUsage.Upload + dbUsage.Download
	return dbUsage, nil
}

func GetAllUsage(db *sql.DB) ([]Usage, error) {
	allUsageQuery := `select date, ssid, upload_usage, download_usage from wifi_usage;`
	dbUsage := []Usage{}
	rows, err := db.Query(allUsageQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var date string
	for rows.Next() {
		var localUsage Usage
		if err = rows.Scan(&localUsage.Date, &localUsage.SSID, &localUsage.Upload, &localUsage.Download); err != nil {
			return nil, err
		}
		fmt.Println("All db date:", date)
		localUsage.TotalUsage = localUsage.Upload + localUsage.Download
		dbUsage = append(dbUsage, localUsage)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return dbUsage, nil
}
func GetUsageByDate(db *sql.DB, date string) ([]Usage, error) {
	allUsageQuery := `select date, ssid, upload_usage, download_usage from wifi_usage where date = ?;`
	dbUsage := []Usage{}
	rows, err := db.Query(allUsageQuery, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var localUsage Usage
		if err = rows.Scan(&localUsage.Date, &localUsage.SSID, &localUsage.Upload, &localUsage.Download); err != nil {
			return nil, err
		}
		localUsage.TotalUsage = localUsage.Upload + localUsage.Download
		dbUsage = append(dbUsage, localUsage)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return dbUsage, nil
}

// Return time in 02-02-2025
func getFormattedTime() string {
	return time.Now().Format("02-01-2006")
}
