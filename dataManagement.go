package main

import (
	"os"
	"encoding/json"
	"io/ioutil"
)

func updateConfigFile(data map[string]interface{}) {
	f, err := os.OpenFile(
		pathToStorageFile,
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
		0666,
	)
	check(err)
	jsonString, err := json.Marshal(data)
	check(err)
	_, err = f.Write(jsonString)
	check(err)
	f.Close()
}

func readFile() (map[string]interface{}, error) {
	//Open file
	data, err := ioutil.ReadFile(pathToStorageFile)
	check(err)

	//Transorm json to map
	err = json.Unmarshal(data, &sshMap)

	return sshMap, err
}

func getMapFromFile() map[string]interface{} {
	sshMap, err := readFile()

	if err == nil {
		return sshMap
	} else {
		return make(map[string]interface{})
	}
}

func createStorageFile() bool {
	//Create file
	_, err := os.Create(pathToStorageFile)
	check(err)
	return true
}

func fileExist(filePath *string) bool {
	if _, err := os.Stat(*filePath); os.IsNotExist(err) {
		return false
	}
	return true
}
