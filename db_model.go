package main

import (
	"encoding/json"
	"log"
	"strconv"
)

type SaveValues struct {
	Year    string     `clover:"year" json:"year"`
	Details []*Details `clover:"details" json:"details"`
	Id      string     `clover:"_id" json:"_id"`
}

type Details struct {
	SSID   string      `clover:"SSID" json:"SSID"`
	Usages []*saveType `clover:"usages" json:"usages"`
}

type saveType struct {
	Date       string `clover:"date" json:"date"`
	TotalUsage uint64 `clover:"usage" json:"usage"`
}

func (r SaveValues) toMap() map[string]any {
	var jsonValue map[string]any
	if err := json.Unmarshal(r.String(), &jsonValue); err != nil {
		log.Fatal(err)
	}
	return jsonValue

}

// / String representation of SaveValues
func (r SaveValues) String() []byte {
	marshal, err := json.Marshal(r)
	if err != nil {
		log.Println(err)
	}
	return marshal
}

func toInt(num string) int {
	converted, err := strconv.Atoi(num)
	if err != nil {
		log.Println(err)
	}
	return converted
}
