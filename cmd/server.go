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

	w.Header().Set("Content-Type", "application/json")

	/*	marshal, err := json.Marshal(allUsage)
		if err != nil {
			http.Error(w, "Error marshalling usage", http.StatusInternalServerError)
		}
		fmt.Println(string(marshal))*/
	err = json.NewEncoder(w).Encode(allUsage)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error converting usage", http.StatusInternalServerError)
	}
	fmt.Println("Returned", allUsage, res.URL.Query().Get("year"))

}

func httpListener() {
	http.HandleFunc("/getUsage", getUsageInfo)
	http.HandleFunc("/getAllUsage", getUsageInfo)
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
