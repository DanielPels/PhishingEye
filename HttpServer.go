package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

//Handle all requests
//-serve files from storage
//-add url's to crawler list

func NewHttpServer(port int) {
	http.HandleFunc("/", httpRequest)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))
}

func httpRequest(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "GET":
		http.ServeFile(writer, request, "screenshots/"+request.URL.Path[1:]+".png")
	case "POST":
		decoder := json.NewDecoder(request.Body)
		var data CrawlRequestPostData
		errorJsonDecode := decoder.Decode(&data)
		if errorJsonDecode != nil {
			writer.WriteHeader(http.StatusBadRequest)
			_, errorHttpWrite := writer.Write([]byte("400 - invalid JSON"))
			if errorHttpWrite != nil {
				log.Println(errorHttpWrite)
			}
			return
		}

		//send binne gekregen data naar channel
		chanCrawlRequestPostData <- &data

		writer.WriteHeader(http.StatusOK)
		_, errorHttpWrite := writer.Write([]byte("200 - OK"))
		if errorHttpWrite != nil {
			log.Println(errorHttpWrite)
		}
	default:
		writer.WriteHeader(http.StatusNotFound)
		_, errorHttpWrite := writer.Write([]byte("404"))
		if errorHttpWrite != nil {
			log.Println(errorHttpWrite)
		}
	}
}
