package main

import (
	"log"
	"os"

	"github.com/nitrous-io/compass"
)

func main() {
	config := &compass.Config{}
	s, err := compass.NewServer(config)
	if err != nil {
		log.Println("Failed to start compass server %v", err)
		os.Exit(1)
		return
	}

	if err := s.ListenAndServe(); err != nil {
		log.Println("compass exit with err %v", err)
		os.Exit(1)
		return
	}
	log.Println("compass exitting....")
}
