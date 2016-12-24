package main

import (
	"encoding/json"
	"log"
	"os"

	"erebus.me/control"
	"erebus.me/model"
)

type Config struct {
	BaseDir    string
	Listen     string
	DataSource string
}

func loadJSONFile(filename string, config interface{}) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	if err := json.NewDecoder(file).Decode(config); err != nil {
		log.Fatal(err)
	}
}

func main() {
	config := Config{}
	loadJSONFile("config.json", &config)

	if err := os.Chdir(config.BaseDir); err != nil {
		log.Fatal(err)
	}

	model.SetDataSource(config.DataSource)

	control.ListenAndServe(config.Listen)
}
