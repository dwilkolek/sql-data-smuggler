package main

import (
	"encoding/json"
	"log"
	"os"
	"path"
	"strings"
)

var example = "example-models"

func main() {
	files, err := os.ReadDir(example)
	if err != nil {
		log.Fatal(err)
	}

	models := make(map[string][]string)

	for _, file := range files {
		if file.IsDir() {
			modelFiles, err := os.ReadDir(path.Join(example, file.Name()))
			if err != nil {
				log.Fatal(err)
			}
			models[file.Name()] = make([]string, 0)
			for _, modelFile := range modelFiles {
				models[file.Name()] = append(models[file.Name()], modelFile.Name())
			}
		} else {
			models[strings.TrimSuffix(file.Name(), ".sql")] = []string{
				file.Name(),
			}
		}
		json, _ := json.Marshal(models)
		log.Default().Printf("%s", json)
	}

}
