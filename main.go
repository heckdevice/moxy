package main

import (
	"fmt"
	"github.com/heckdevice/moxy/core"
	log "github.com/sirupsen/logrus"
	"os"
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetOutput(os.Stderr)
	log.SetFormatter(&log.JSONFormatter{})
	service, err := core.RegisterService("test_service", "1.0")
	if err != nil {
		log.Error(fmt.Sprintf("Error registering services :: %v", err))
	} else {
		log.Info(fmt.Sprintf("%v registered", service))
		api, err := service.RegisterAPI("/helloworld", core.GET)
		if err != nil {
			log.Error(fmt.Sprintf("Error registering api :: %v", err))
		} else {
			log.Info(fmt.Sprintf("%v registered with %v", api, service))
		}
		apiLatency, err := service.RegisterAPIWithLatency("/helloworldagain", core.GET, 10.2)
		if err != nil {
			log.Error(fmt.Sprintf("Error registering api :: %v", err))
		} else {
			log.Info(fmt.Sprintf("%v registered with %v", apiLatency, service))
		}
	}
}
