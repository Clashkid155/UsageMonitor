package main

import (
	"encoding/json"
	"fmt"
	"github.com/ostafen/clover/v2"
	"log"
	"net/http"
)

func getUsageInfo(w http.ResponseWriter, res *http.Request) {
	fmt.Println(res.Body)
	year := res.FormValue("year") //PostForm.Get("year")
	all, err := cDB.FindFirst(clover.NewQuery(collectionName).Where(clover.Field("year").Eq(year)))
	if err != nil {
		log.Println(err)
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(all.ToMap())
	if err != nil {
		log.Println(err)
	}
	fmt.Println("Returned", all.ToMap(), res.URL.Query().Get("year"))
}

func httpListener() {
	http.HandleFunc("/getUsage", getUsageInfo)
	err := http.ListenAndServe("127.0.0.1:9083", nil)
	if err != nil {
		log.Println(err)
	}
}
