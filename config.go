package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Ignored []string
}

func (c Config) Save() error {

	// Convert the File object to a Json object
	file, err := os.Create(ATHINA_CONFIG)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(c)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (c Config) IsIgnored(filename string) bool {

	for _, ignored := range c.Ignored {
		if ignored == filename {
			return true
		}
	}

	return false
}
