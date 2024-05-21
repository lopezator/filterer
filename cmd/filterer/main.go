package main

import (
	"log"
	"os"

	"github.com/lopezator/filterer/internal/server"
	"gopkg.in/yaml.v3"
)

func main() {
	// Read the config file.
	data, err := os.ReadFile("filterer.yaml")
	if err != nil {
		log.Fatalf("filterer: error reading filterer.yaml config file: %s\n", err)
	}

	// Define a variable of type Config.
	var config server.Config

	// Unmarshal the YAML data into the Config variable.
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("Error parsing YAML file: %s\n", err)
	}

	// Run the filterer server.
	srv, err := server.New(&server.Config{
		Addr:      config.Addr,
		FieldSets: config.FieldSets,
	})
	if err != nil {
		log.Fatal(err)
	}
	if err := srv.Serve(); err != nil {
		log.Fatal(err)
	}
}
