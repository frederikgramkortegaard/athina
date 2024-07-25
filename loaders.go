package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func loadAthinaFileObject(identifier string) (AthinaFile, error) {

	// In '.athina/objects', there should be a file with the name of the identifier
	// If the file does not exist, return an error, otherwise we can load the file in as a File object
	if _, err := os.Stat(".athina/objects/" + identifier); os.IsNotExist(err) {
		return AthinaFile{}, err
	}

	// Load it as a Json object into a File struct
	file, err := os.Open(".athina/objects/" + identifier)
	if err != nil {
		fmt.Println(err)
		return AthinaFile{}, err
	}
	defer file.Close()

	var f AthinaFile
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&f)
	if err != nil {
		fmt.Println(err)
		return AthinaFile{}, err
	}

	return f, nil

}

func loadConfig() {
	file, _ := os.Open(ATHINA_CONFIG)
	defer file.Close()
	decoder := json.NewDecoder(file)
	_ = decoder.Decode(&config)
}
