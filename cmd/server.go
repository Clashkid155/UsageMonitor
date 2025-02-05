package main

import (
	"encoding/json"
	"fmt"
	"github.com/clashkid155/usage-monitor"
	"log"
	"net/http"
)

func getUsageInfo(w http.ResponseWriter, res *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	allUsage, err := usageTracker.GetAllUsage(sqlDb)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusForbidden)
		_ = response(&JsonResponse{
			Message: "failed",
			Error:   err.Error(),
		}, w)
		return
	}

	err = json.NewEncoder(w).Encode(allUsage)
	err = response(&JsonResponse{
		Message: "success",
		Data:    allUsage,
	}, w)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		_ = response(&JsonResponse{
			Message: "failed",
			Error:   err.Error(),
		}, w)
		return
	}
	fmt.Println("Returned", allUsage, res.URL.Query().Get("year"))

}

func getUsageByDate(w http.ResponseWriter, res *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	year := res.URL.Query().Get("year")
	if year == "" {
		w.WriteHeader(http.StatusForbidden)
		_ = response(&JsonResponse{
			Message: "failed",
			Error:   "missing year parameter",
		}, w)

		return
	}
	allUsage, err := usageTracker.GetUsageByDate(sqlDb, year)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		_ = response(&JsonResponse{
			Message: "failed",
			Error:   err.Error(),
		}, w)
		// http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// err = json.NewEncoder(w).Encode(allUsage)
	err = response(&JsonResponse{
		Message: "success",
		Data:    allUsage,
	}, w)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		_ = response(&JsonResponse{
			Message: "failed",
			Error:   err.Error(),
		}, w)
		return
	}
	fmt.Println("Returned", allUsage, res.URL.Query().Get("year"))

}

func httpListener() {
	http.HandleFunc("/getAllUsage", getUsageInfo)
	http.HandleFunc("/getUsageByDate", getUsageByDate)
	err := http.ListenAndServe(":9083", nil)
	if err != nil {
		log.Println(err)
	}
}

// {"message":"Successful",
// "data":[],
// "error":"no row"}

type JsonResponse struct {
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

func response(jsonRes *JsonResponse, w http.ResponseWriter) error {
	err := json.NewEncoder(w).Encode(jsonRes)
	if err != nil {
		return err
	}
	return nil
}
