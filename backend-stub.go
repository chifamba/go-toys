package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"golang/yangu/certs"
	"github.com/gorilla/mux"
)

const MaxContentLenght = 1024*1024

var client *http.Client

var BE_PROXY_URL string
var BE_HOSTNAME string
var PROXY_ENABLE bool

type ResponseMsg struct {
	SRC_HOST []string `json:"remoteHost,omitempty"`
	SERVER []string `json:"serverHost,omitempty"`
	Message string `json:"message,omitempty"`
	APIVersion string `json:"apiVersion,omitempty"`
	START_TIMESTAMP time.Time `json:"startTimeStamp,omitempty"`
	END_TIMESTAMP time.Time `json:"endTimeStamp,omitempty"`
}

func init()  {
/*
Initialise the application
*/
	be_host,err:=os.Hostname()
	if err!=nil {
		log.Println("Failed to determine hostname, I'm going to use localhost now, hope thats OK.")
		BE_HOSTNAME="localhost"
	} else {
		BE_HOSTNAME=be_host
	}

	//get all the dynamic config and stuff..whatever
	bootstrap()

	//get the http client ready
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client = &http.Client{Transport: tr}

	//schedule regular refesh of the config..
	schedule(bootstrap, 1*time.Minute)
}

func bootstrap(){
	/*
	Load/reload config..
	 */

	log.Println("Refreshing config..")
	url:=os.Getenv("BE_PROXY_URL")

	if url!=BE_PROXY_URL {
		if len(url)>9 {
			BE_PROXY_URL=url
			PROXY_ENABLE=true
			log.Println("Setting BE_PROXY_URL [",BE_PROXY_URL,"]")

		} else {
			BE_PROXY_URL="NULL"
			PROXY_ENABLE=false
			log.Println(" PROXY Disabled!")
		}
	}
}


func schedule(what func(), delay time.Duration) chan bool {
	/*
	scheduler
	*/
	stop := make(chan bool)
	go func() {
		for {
			what()
			select {
			case <-time.After(delay):
			case <-stop:
				return
			}
		}
	}()
	return stop
}



// http route handlers
func DefaultGet(res http.ResponseWriter,req *http.Request){
	resMsg:=ResponseMsg{
		Message:"Hello there, thanks for stopping by..",
		APIVersion:req.Header.Get("X-API-VERSION"),
		START_TIMESTAMP:time.Now(),
		END_TIMESTAMP:time.Now(),
	}

	resMsg.SRC_HOST=append(resMsg.SRC_HOST,req.Host)
	resMsg.SERVER=append(resMsg.SERVER,BE_HOSTNAME)

	if PROXY_ENABLE {
		res.Write(makeRequest(resMsg))
	} else {
		res.WriteHeader(http.StatusOK)
		responseJson,err:=json.MarshalIndent(resMsg,"","\t")
		if err!=nil {
			log.Println("Failed to marshal response ..")
			res.Write([]byte("OK"))
		} else {
			res.Write(responseJson)
		}
	}

}


func ProxyBackEnd(res http.ResponseWriter,req *http.Request){

	// limit the content legth to
	req.Body = http.MaxBytesReader(res, req.Body, MaxContentLenght)
	decoder:=json.NewDecoder(req.Body)
	var resMsg ResponseMsg
	err:=decoder.Decode(&resMsg)
	if err != nil {
		log.Println(err)
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	resMsg.SRC_HOST=append(resMsg.SRC_HOST,req.Host)
	resMsg.SERVER=append(resMsg.SERVER,BE_HOSTNAME)
	resMsg.END_TIMESTAMP=time.Now()
	res.WriteHeader(http.StatusOK)

	responseJson,err:=json.MarshalIndent(resMsg,"","\t")
	if err!=nil {
		log.Println("Failed to marshal response ..")
		res.Write([]byte("OK"))
	} else {
		res.Write(responseJson)
	}
}

func startApiHandlers()  {
	//create http router using mux
	router :=mux.NewRouter()

	//set CA chains --using self signed certificates
	caCertRoot, err := ioutil.ReadFile("cert.pem")

	if err!=nil {
		log.Fatal(err)
	}

	caCertPoolRoot := x509.NewCertPool()
	caCertPoolRoot.AppendCertsFromPEM(caCertRoot)

	// set http server options
	server := &http.Server{
		Addr: ":8080",
		Handler:router,
		TLSConfig: &tls.Config{
			ClientCAs:caCertPoolRoot,
			ClientAuth: tls.NoClientCert,
			InsecureSkipVerify: true,
		},
	}

	//add routes
	go  router.HandleFunc("/",DefaultGet).Methods("GET")
	go  router.HandleFunc("/proxy",ProxyBackEnd).Methods("POST")

	//start http server
	log.Println(" === Starting server on port [",server.Addr,"]")
	log.Fatal(server.ListenAndServeTLS("cert.pem", "key.pem"))
}


func makeRequest(req_data ResponseMsg) []byte  {

	jsonStrRequest,_:=json.Marshal(req_data)
	req, err := http.NewRequest("POST", BE_PROXY_URL, bytes.NewBuffer(jsonStrRequest))
	if err!=nil {
		log.Println("Failed to create a request..")
	}

	req.Header.Set("Accepts-version", "1.0")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err, " something went wrong. Is the server up?")
		return nil
	} else  {
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)
		log.Println("Response Status:", resp.Status, " Response Headers:", resp.Header," Response Body:", string(body))
		return body
	}
}

/*
Toy go code...just messing around, nothing serious.
This code will only work at exactly "2009-11-10 23:00:00 +0000 UTC m=+0.000000001"
*/

func main(){

	log.Println(" === Checking Certificates..")
	certs.CheckCerts()
	log.Println(" === Starting Server.")
	go  startApiHandlers()
	time.Sleep(2000)
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(<-ch)
}
