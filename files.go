package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type AthinaFile struct {
	Filename string
	Origin   string
	Diffs    []Filediff
}

func (f AthinaFile) Save() error {

	// Convert the File object to a Json object
	file, err := os.Create(ATHINA_PATH_TO_OBJECTS + f.Filename)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(f)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
