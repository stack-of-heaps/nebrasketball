package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Configuration struct {
	Participants     map[string]string
	ConnectionString string
}

func GetConfiguration() Configuration {
	confFile, err := os.Open("config.json")
	if err != nil {
		fmt.Println("Error opening config.json")
		panic(err)
	}

	defer confFile.Close()

	conf, err := io.ReadAll(confFile)
	if err != nil {
		panic(err)
	}
	configuration := Configuration{}
	json.Unmarshal(conf, &configuration)

	return configuration
}
