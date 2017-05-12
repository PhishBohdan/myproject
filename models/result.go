package models

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/jinzhu/gorm"
	"github.com/oschwald/maxminddb-golang"
)

type mmCity struct {
	GeoPoint mmGeoPoint `maxminddb:"location"`
}

type mmGeoPoint struct {
	Latitude  float64 `maxminddb:"latitude"`
	Longitude float64 `maxminddb:"longitude"`
}

// Result contains the fields for a result object,
// which is a representation of a target in a campaign.
type Result struct {
	Id         int64   `json:"-"`
	CampaignId int64   `json:"campaignid"`
	UserId     int64   `json:"-"`
	RId        string  `json:"id"`
	Email      string  `json:"email"`
	FirstName  string  `json:"first_name"`
	LastName   string  `json:"last_name"`
	Position   string  `json:"position"`
	Department string  `json:"department"`
	Status     string  `json:"status" sql:"not null"`
	IP         string  `json:"ip"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
}

// TrimmedResult contains the fields for a result object,
// which is a representation of a target in a campaign.
type TrimmedResult struct {
	CampaignID   int64  `json:"campaignid"`
	CampaignName string `json:"campaignname"`
	RId          string `json:"id"`
	Email        string `json:"email"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Position     string `json:"position"`
	Department   string `json:"department"`
	Status       string `json:"status" sql:"not null"`
}

// UpdateStatus updates the status of the result in the database
func (r *Result) UpdateStatus(s string) error {
	return db.Table("results").Where("id=?", r.Id).Update("status", s).Error
}

// UpdateGeo updates the latitude and longitude of the result in
// the database given an IP address
func (r *Result) UpdateGeo(addr string) error {
	// Open a connection to the maxmind db
	mmdb, err := maxminddb.Open("static/db/geolite2-city.mmdb")
	if err != nil {
		log.Fatal(err)
	}
	defer mmdb.Close()
	ip := net.ParseIP(addr)
	var city mmCity
	// Get the record
	err = mmdb.Lookup(ip, &city)
	if err != nil {
		return err
	}
	// Update the database with the record information
	return db.Table("results").Where("id=?", r.Id).Updates(map[string]interface{}{
		"ip":        addr,
		"latitude":  city.GeoPoint.Latitude,
		"longitude": city.GeoPoint.Longitude,
	}).Error
}

// GenerateId generates a unique key to represent the result
// in the database
func (r *Result) GenerateId() {
	// Keep trying until we generate a unique key (shouldn't take more than one or two iterations)
	k := make([]byte, 32)
	for {
		io.ReadFull(rand.Reader, k)
		r.RId = fmt.Sprintf("%x", k)
		err := db.Table("results").Where("r_id=?", r.RId).First(&Result{}).Error
		if err == gorm.ErrRecordNotFound {
			break
		}
	}
}

// GetResult returns the Result object from the database
// given the ResultId
func GetResult(rid string) (Result, error) {
	r := Result{}
	err := db.Where("r_id=?", rid).First(&r).Error
	return r, err
}
