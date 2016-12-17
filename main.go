package main

import (
	"encoding/json"
	"log"
	"os"
	"path"

	"erebus.me/control"
	"erebus.me/model"
	"erebus.me/view"
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

	templateDir := path.Join(config.BaseDir, "/template")
	view.SetTemplateDir(templateDir)

	documentRoot := path.Join(config.BaseDir, "/html")
	view.SetDocumentRoot(documentRoot)

	control.ListenAndServe(config.Listen)
}
