package main

import (
	"context"
	"fmt"
	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/network"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/mafredri/cdp/rpcc"
	"log"
	"strconv"
	"time"
)

type CrawlRequest struct {
	chromeInstance *ChromeInstance
	client         *cdp.Client
	conn           *rpcc.Conn
	ctx            context.Context
	cancel         context.CancelFunc
	URL            string
}

//URL Slice

//Crawl een URL die in de lijst staat
//haal een chrome instance op
//maak connectie met chrome via CDP
//navigeer naar de website
//maak screenshot als er geen activiteit is voor 10 sec
//sla alles op via storage

func CrawlURL(uuid, URL string, port int, done chan int) {
	log.Println("New crawl request:", uuid, URL)

	chromeInstance := NewChromeInstance(port, uuid)
	err := chromeInstance.Start()
	if err != nil {
		log.Println(uuid, err)
		done <- port
		return
	}
	time.Sleep(time.Second * 1)

	c := &CrawlRequest{
		chromeInstance: chromeInstance,
		client:         nil,
		conn:           nil,
		ctx:            nil,
		cancel:         nil,
		URL:            URL,
	}
	c.ctx, c.cancel = context.WithCancel(context.Background())
	defer c.cancel()

	err = createConnectionToBrowser(c, chromeInstance.GetPort())
	if err != nil {
		log.Println(uuid, err)
		done <- port
		return
	}

	// Open a DOMContentEventFired client to buffer this event.
	domContent, err := c.client.Page.DOMContentEventFired(c.ctx)
	if err != nil {
		log.Println(uuid, err)
		done <- port
		return
	}
	defer domContent.Close()

	err = c.client.Page.Enable(c.ctx)
	if err != nil {
		log.Println(uuid, err)
		done <- port
		return
	}

	err = c.client.Network.Enable(c.ctx, network.NewEnableArgs())
	if err != nil {
		log.Println(uuid, err)
		done <- port
		return
	}

	err = c.client.Network.ClearBrowserCache(c.ctx)
	if err != nil {
		log.Println(uuid, err)
		done <- port
		return
	}

	err = c.client.Network.SetCacheDisabled(c.ctx, network.NewSetCacheDisabledArgs(true))
	if err != nil {
		log.Println(uuid, err)
		done <- port
		return
	}

	requestPatterns := make([]network.RequestPattern, 0)
	searchString := "*"
	requestPatterns = append(requestPatterns, network.RequestPattern{
		URLPattern:        &searchString,
		ResourceType:      "",
		InterceptionStage: "",
	})
	err = c.client.Network.SetRequestInterception(c.ctx, network.NewSetRequestInterceptionArgs(requestPatterns))
	if err != nil {
		log.Println(uuid, err)
		done <- port
		return
	}

	interceptedClient, err := c.client.Network.RequestIntercepted(c.ctx)
	if err != nil {
		log.Println(uuid, err)
		done <- port
		return
	}

	responseReceivedClient, err := c.client.Network.ResponseReceived(c.ctx)
	if err != nil {
		log.Println(uuid, err)
		done <- port
		return
	}

	go func(ctx context.Context) {
		for {
			select {
			case <-interceptedClient.Ready():
				reply, errReply := interceptedClient.Recv()
				if errReply != nil {
					fmt.Println("intercept recv:", errReply)
					continue
				}
				interceptedRequestArgs := network.NewContinueInterceptedRequestArgs(reply.InterceptionID)
				errIntercept := c.client.Network.ContinueInterceptedRequest(ctx, interceptedRequestArgs)
				if errIntercept != nil {
					fmt.Println("intercept accept", errIntercept)
				}
				continue
			case <-responseReceivedClient.Ready():
				_, errReceived := responseReceivedClient.Recv()
				if errReceived != nil {
					fmt.Println("received err", errReceived)
				}
				continue
			case <-ctx.Done():
				return
			}
		}
	}(c.ctx)

	navArgs := page.NewNavigateArgs("https://goo.gl/W3mgAf")
	_, err = c.client.Page.Navigate(c.ctx, navArgs)
	if err != nil {
		log.Println(uuid, err)
		done <- port
		return
	}

	// Wait until we have a DOMContentEventFired event.
	if _, err = domContent.Recv(); err != nil {
		log.Println(uuid, err)
		done <- port
		return
	}

	//nu een domme slaap ipv iets slims dat checked of de site geladen is.
	time.Sleep(time.Second * 15)

	screenshot, err := c.client.Page.CaptureScreenshot(c.ctx, page.NewCaptureScreenshotArgs().SetFormat("png").SetQuality(100))
	if err != nil {
		log.Println(uuid, err)
		done <- port
		return
	}

	//screenshot is gemaakt verstuur naar done channel
	chanCrawlCompleteData <- &CrawlCompleteData{
		UUID:       uuid,
		ScreenShot: screenshot.Data,
	}

	//cleanup
	err = c.client.Browser.Close(c.ctx)
	if err != nil {
		log.Println(uuid, err)
		done <- port
		return
	}

	err = c.conn.Close()
	if err != nil {
		log.Println(uuid, err)
		done <- port
		return
	}

	//alles zou goed moeten gegaan zijn!
	done <- port
}

func createConnectionToBrowser(crawlRequest *CrawlRequest, port int) error {
	// Use the DevTools HTTP/JSON API to manage targets (e.g. pages, webworkers).
	devt := devtool.New("http://127.0.0.1:" + strconv.Itoa(port))
	pt, err := devt.Get(crawlRequest.ctx, devtool.Page)
	if err != nil {
		pt, err = devt.Create(crawlRequest.ctx)
		if err != nil {
			return err
		}
	}

	// Initiate a new RPC connection to the Chrome DevTools Protocol target.
	conn, err := rpcc.DialContext(crawlRequest.ctx, pt.WebSocketDebuggerURL)
	if err != nil {
		return err
	}

	crawlRequest.conn = conn
	crawlRequest.client = cdp.NewClient(conn)

	return nil
}
