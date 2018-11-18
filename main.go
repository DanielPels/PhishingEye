package main

import (
	"log"
	"sync"
)

type CrawlRequestPostData struct {
	URL  string `json:"url"`
	UUID string `json:"uuid"`
}

type CrawlCompleteData struct {
	UUID       string
	ScreenShot []byte
}

var chanCrawlRequestPostData chan *CrawlRequestPostData
var chanCrawlCompleteData chan *CrawlCompleteData
var chanWorkerDone chan int

var urlRequest mutexSlice

//totaal aantal chrome browsers die open mogen
var totalWorkers mutexCounter

var NumberPool numberPool

func main() {
	log.Println("Starting PhishingEye")
	//setup url request slice
	urlRequest = mutexSlice{
		mu:    sync.Mutex{},
		slice: make([]*CrawlRequestPostData, 0),
	}

	NumberPool = numberPool{
		mu:   sync.Mutex{},
		pool: []int{},
	}

	for i := 0; i < 20; i++ {
		NumberPool.ReleaseNumber(i + 9000)
	}

	//setup nieuw chan waar alle post data komt van de http server
	chanCrawlRequestPostData = make(chan *CrawlRequestPostData, 100)
	chanCrawlCompleteData = make(chan *CrawlCompleteData, 100)
	chanWorkerDone = make(chan int, 20)
	totalWorkers = mutexCounter{
		mu: sync.Mutex{},
		x:  1,
	}

	go func() {
		for {
			select {
			case data := <-chanCrawlRequestPostData:
				if totalWorkers.Value() > 0 && NumberPool.IsNumberAvailable() {
					newCrawlFromHttp(data, NumberPool.GetNumber())
				} else {
					//voeg aan url request slice toe
					urlRequest.AddData(data)
				}
			case data := <-chanCrawlCompleteData:
				go SaveCrawlerData(data)
			case number := <-chanWorkerDone:
				totalWorkers.Add(1)
				NumberPool.ReleaseNumber(number)
				if urlRequest.Len() > 0 {
					if NumberPool.IsNumberAvailable() {
						newCrawlFromHttp(urlRequest.GetFirstElement(), NumberPool.GetNumber())
					}
				}
			}
		}
	}()
	NewHttpServer(3001)
}

func newCrawlFromHttp(data *CrawlRequestPostData, port int) {
	totalWorkers.Subtract(1)
	go CrawlURL(data.UUID, data.URL, port, chanWorkerDone)
}

type mutexCounter struct {
	mu sync.Mutex
	x  int
}

func (c *mutexCounter) Subtract(x int) {
	c.mu.Lock()
	c.x -= x
	c.mu.Unlock()
}

func (c *mutexCounter) Add(x int) {
	c.mu.Lock()
	c.x += x
	c.mu.Unlock()
}

func (c *mutexCounter) Value() int {
	return c.x
}

type mutexSlice struct {
	mu    sync.Mutex
	slice []*CrawlRequestPostData
}

func (s *mutexSlice) AddData(data *CrawlRequestPostData) {
	s.mu.Lock()
	s.slice = append(s.slice, data)
	s.mu.Unlock()
}

func (s *mutexSlice) GetFirstElement() *CrawlRequestPostData {
	s.mu.Lock()
	returnData := s.slice[0]
	s.slice[0] = nil
	s.slice = s.slice[1:]
	s.mu.Unlock()
	return returnData
}

func (s *mutexSlice) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.slice)
}

type numberPool struct {
	mu   sync.Mutex
	pool []int
}

func (n *numberPool) GetNumber() int {
	n.mu.Lock()
	returnData := n.pool[0]
	n.pool = n.pool[1:]
	n.mu.Unlock()
	return returnData
}

func (n *numberPool) ReleaseNumber(i int) {
	n.mu.Lock()
	n.pool = append(n.pool, i)
	n.mu.Unlock()
}

func (n *numberPool) IsNumberAvailable() bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	return len(n.pool) > 0
}
