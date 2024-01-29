package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Configuration struct {
	Participants  map[string]string
	Db            Db
	ServerAddress string
}

type Db struct {
	Collection       string
	ConnectionString string
	Database         string
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
