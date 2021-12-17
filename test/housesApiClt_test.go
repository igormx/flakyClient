package main

import (
	"flakyClient/services"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type HTTPMockClient struct {
	mock.Mock
}

func TestDownloadHouseImage(t *testing.T) {
	//Crating mock http client
	fileContent := io.NopCloser(strings.NewReader("file content"))
	mockResponse := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       fileContent,
	}
	httpMockClient := new(HTTPMockClient)
	httpMockClient.On("Do", mock.Anything).Return(mockResponse, nil)

	//Api client
	housesApiClient := services.HousesApiClient{
		URLBase: "https://app-homevision-staging.herokuapp.com/api_project/houses",
		HTTPClient: &http.Client{
			Timeout: time.Minute,
		},
		MaxRetries: 4,
	}

	//Invoke actual func
	wg := &sync.WaitGroup{}
	errorChan := make(chan error)
	url := "http://mocksite.com/mockimage.jpg"
	id := 33
	formatedAddress := services.FormatAddress("838 James Rd, Irving Tx.")
	fileName := "id-" + strconv.Itoa(int(id)) + "-" + formatedAddress + filepath.Ext(url)
	wg.Add(1)
	housesApiClient.DownloadHouseImage(url, 33, formatedAddress, errorChan, wg)

	//Assertion, file has been "downloaded"?
	assert.FileExists(t, fileName)
	if _, err := os.Stat(fileName); err == nil {
		//remove file if exists
		os.Remove(fileName)
	}
}
