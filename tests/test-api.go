package main

import (
	"net/http"
	"bytes"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"
	"crypto/tls"
	"math/rand"
	"encoding/json"
	"log"
)

/*
while true;
	do
		curl -k -H "Accepts-version:1.0" -X POST -d '{"translation":"NIV","book":"Psalms","chapter":"121","verse":"2"}' https://localhost:8080/canonical/json ;
		echo;
		sleep 10;
	done;
*/
var availBooks []string
var availLang []string
var client *http.Client

type rx struct {
	SRC_HOST []string `json:"remoteHost,omitempty"`
	SERVER []string `json:"serverHost,omitempty"`
	Message string `json:"message,omitempty"`
	APIVersion string `json:"apiVersion,omitempty"`
	START_TIMESTAMP time.Time `json:"startTimeStamp,omitempty"`
	END_TIMESTAMP time.Time `json:"endTimeStamp,omitempty"`
}

func init()  {

}

func workerThread()  {
	for {

		//handleRequest(makeVerseRequest())
		handleRequest(makeVOTDRequest())
		time.Sleep(time.Millisecond * 500)
	}
}

func main()  {
	//availBooks=[]string{"ESV","MSG","NIV","NLT"}
	//availLang=[]string{"EN"}

	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client = &http.Client{Transport: tr}

	for i := 1; i <= 10; i++ {
		go  workerThread()
	}
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(<-ch)

}

func handleRequest(req *http.Request)  {
	if req==nil {
		log.Println("NIL Request! ")
	} else {
		resp, err := client.Do(req)
		if err != nil {
			log.Println(err, " pausing for a bit and will try again. Is the server up?")
			time.Sleep(time.Second * 10)
		} else  {
			defer resp.Body.Close()

			body, _ := ioutil.ReadAll(resp.Body)
			log.Println("Response Status:", resp.Status, " Response Headers:", resp.Header," Response Body:", string(body))

		}

		//os.Exit(0)
	}
}

func random(min, max int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	return rand.Intn(max - min) + min
}

func makeVOTDRequest() *http.Request {
	url := "https://192.168.64.10:32562"

	var jSRaw rx


	//var jsonStrRequest = []byte(jSRaw)
	jsonStrRequest,_:=json.Marshal(jSRaw)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStrRequest))
	if err!=nil {
		return nil
	}

	req.Header.Set("Accepts-version", "1.0")
	req.Header.Set("Content-Type", "application/json")
	return req
}