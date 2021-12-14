package main

import "fmt"

func main() {
	housesApiClient := NewHousesApiClient()
	housesInfo, err := housesApiClient.FetchHousesInfoPage(1)

	fmt.Printf("%v", housesInfo)
	fmt.Println(err)
}
