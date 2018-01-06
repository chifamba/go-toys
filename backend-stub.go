package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main(){

	log.Println(" === Checking Certificates..")
	log.Println(" === Starting Server.")
	time.Sleep(2000)
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(<-ch)
}
