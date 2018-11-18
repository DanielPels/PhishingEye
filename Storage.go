package main

import (
	"io/ioutil"
	"log"
	"os"
)

//Save data from UUID
func SaveCrawlerData(data *CrawlCompleteData) {
	if _, err := os.Stat("./screenshots"); os.IsNotExist(err) {
		os.Mkdir("./screenshots", 0600)
	}

	err := ioutil.WriteFile("screenshots/"+data.UUID+".png", data.ScreenShot, 0600)
	if err != nil {
		log.Println(err)
	}
}

//Get data from UUID
func GetCrawlDataWithUUID(uuid string) ([]byte, error) {
	return ioutil.ReadFile(uuid + ".png")
}
