package services

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type HousesInfo struct {
	Houses []HouseDetail `json:"houses"`
	Ok     bool          `json:"ok"`
	Error  error
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
		MaxRetries: 4,
	}
}

func (housesClient *HousesApiClient) GetHousesImages() {
	wg := &sync.WaitGroup{}
	housesChan := make(chan HousesInfo)
	errorChan := make(chan error)

	go func() {
		wg.Wait()
		close(housesChan)
		close(errorChan)
	}()

	for page := 1; page <= 10; page++ {
		wg.Add(1)
		go housesClient.GetHousesInfoPage(page, housesChan, errorChan, wg)
	}

	for housesInfo := range housesChan {
		if housesInfo.Error != nil {
			log.Fatal("Failing to get houses detail after "+strconv.Itoa(housesClient.MaxRetries)+" retries. Error: ", housesInfo.Error)
		} else {
			for _, house := range housesInfo.Houses {
				wg.Add(1)
				go housesClient.DownloadHouseImage(house.PhotoURL, house.ID, house.Address, errorChan, wg)
			}
		}
	}

	select {
	case <-errorChan:
		break
	case err := <-errorChan:
		{
			if err != nil {
				log.Println("Error during the image download: ", err)
			}
		}
	}

	log.Println("Success")
}

func (housesClient *HousesApiClient) GetHousesInfoPage(page int, housesChan chan HousesInfo, errorChan chan error, wg *sync.WaitGroup) {
	defer wg.Done()

	retries := 1
	housesInfo := HousesInfo{}
	for retries <= housesClient.MaxRetries {
		log.Println("Fetching Houses Info Page " + strconv.Itoa(page) + ", try " + strconv.Itoa(retries) + "...")

		request, err := http.NewRequest("GET", housesClient.URLBase+"?page="+strconv.Itoa(page), nil)
		if err != nil {
			housesInfo.Error = err
			retries++
			continue
		}

		response, err := housesClient.HTTPClient.Do(request)
		if err != nil {
			housesInfo.Error = err
			retries++
			continue
		}

		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			housesInfo.Error = errors.New("invalid HTTP Response Code")
			retries++
			continue
		}

		body, _ := ioutil.ReadAll(response.Body)
		err = json.Unmarshal(body, &housesInfo)
		if err != nil {
			housesInfo.Error = err
			retries++
			continue
		}
		housesInfo.Error = nil
		break
	}
	housesChan <- housesInfo
}

func (housesClient *HousesApiClient) DownloadHouseImage(url string, id int64, address string, errorChannel chan error, wg *sync.WaitGroup) {
	defer wg.Done()
	fileName := "id-" + strconv.Itoa(int(id)) + "-" + FormatAddress(address) + filepath.Ext(url)
	if _, err := os.Stat(fileName); err == nil {
		log.Println("Skipping ", fileName, " already exists")
		return
	}
	log.Println("Downloading ", fileName)

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		errorChannel <- err
	}
	resp, err := housesClient.HTTPClient.Do(request)
	if err != nil {
		errorChannel <- err
	}
	defer resp.Body.Close()

	file, err := os.Create(fileName)
	if err != nil {
		errorChannel <- err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		errorChannel <- err
	}
	log.Println(file.Name(), " downloaded")
}

func FormatAddress(address string) string {
	address = strings.ReplaceAll(address, " ", "_")
	address = strings.ReplaceAll(address, ".", "")
	address = strings.ReplaceAll(address, ",", "")
	address = strings.ReplaceAll(address, ";", "")
	return strings.ToLower(address)
}
