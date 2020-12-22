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
		payload := core.Payload{"test": "complete", "test2": "complete"}
		api, err := service.RegisterAPI("/helloworld", core.GET, payload)
		if err != nil {
			log.Error(fmt.Sprintf("Error registering api :: %v", err))
		} else {
			log.Info(fmt.Sprintf("%v registered with %v", api, service))
		}
		payload = core.Payload{"test2": "complete", "test": "complete"}

		_, err = service.RegisterAPI("/helloworld", core.GET, payload)
		if err != nil {
			log.Error(fmt.Sprintf("Error registring the api :: %v", err))
		}
		apiLatency, err := service.RegisterAPIWithLatency("/slowhelloworld", core.GET, nil, 10.2)
		if err != nil {
			log.Error(fmt.Sprintf("Error registering api :: %v", err))
		} else {
			log.Info(fmt.Sprintf("%v registered with %v", apiLatency, service))
		}
	}
}
