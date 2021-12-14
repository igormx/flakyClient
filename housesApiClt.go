package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

type HousesInfo struct {
	Houses []HouseDetail `json:"houses"`
	Ok     bool          `json:"ok"`
}

type HouseDetail struct {
	ID        int64  `json:"id"`
	Address   string `json:"address"`
	Homeowner string `json:"homeowner"`
	Price     int64  `json:"price"`
	PhotoURL  string `json:"photoURL"`
}

type HousesApiClient struct {
	URLBase    string
	HTTPClient *http.Client
	MaxRetries int
}

func NewHousesApiClient() *HousesApiClient {
	return &HousesApiClient{
		URLBase: "https://app-homevision-staging.herokuapp.com/api_project/houses",
		HTTPClient: &http.Client{
			Timeout: time.Minute,
		},
		MaxRetries: 3,
	}
}

func (housesClient *HousesApiClient) FetchHousesInfoPage(page int) (HousesInfo, error) {
	var retries int
	var returnError error
	retries = 1

	for retries <= housesClient.MaxRetries {
		log.Println("Fetching Houses info, try " + strconv.Itoa(retries) + "...")

		request, err := http.NewRequest("GET", housesClient.URLBase+"?page="+strconv.Itoa(page), nil)
		if err != nil {
			log.Println("Error during request creation ", err)
			returnError = err
			retries++
			continue
		}

		response, err := housesClient.HTTPClient.Do(request)
		if err != nil {
			log.Println("Error, server side ", err)
			returnError = err
			retries++
			continue
		}

		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			log.Println("Error, invalid HTTP Response Code")
			returnError = errors.New("invalid HTTP Response Code")
			retries++
			continue
		}

		housesInfo := HousesInfo{}
		body, _ := ioutil.ReadAll(response.Body)
		err = json.Unmarshal(body, &housesInfo)
		if err != nil {
			log.Println("Error, invalid Response Content")
			returnError = err
			retries++
			continue
		}

		return housesInfo, nil
	}

	return HousesInfo{}, returnError
}
