package main

import "flakyClient/services"

func main() {
	housesApiClient := services.NewHousesApiClient()
	housesApiClient.GetHousesImages()
}
