package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/nitrous-io/compass"
)

func main() {
	configData, err := ioutil.ReadFile("config.json")
	if err != nil {
		fatal("Missing config file at ./config.json")
	}
	config := new(compass.Config)
	err = json.Unmarshal(configData, &config)
	if err != nil {
		fatal("Invalid config file at ./config.json", err)
	}
	s, err := compass.NewServer(config)
	if err != nil {
		fatal("Failed to start compass server %v", err)
	}
	log.Println("Serving ...")
	if err := s.ListenAndServe(); err != nil {
		fatal("compass exit with err %v", err)
	}
}

func fatal(a ...interface{}) {
	log.Println(a...)
	os.Exit(1)
}
