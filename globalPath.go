package main

import (
	"github.com/mitchellh/mapstructure"
)

func createFromMap(m map[string]interface{}) (Host) {
	var result Host
	mapstructure.Decode(m, &result)
	return result
}

func updateGlobalPath(key string) {
	globalPath = append(globalPath, key)
}

func getInterfaceFromPaths(m interface{}, paths []string) interface{} {
	inter := m
	for _, path := range paths{
		inter = inter.(map[string]interface{})[path]
	}
	return inter
}

func getMapAtGlobalPath() interface{} {
	return getInterfaceFromPaths(sshMap, globalPath)
}

func changeToUpperLevelProject() {
	if len(globalPath) > 0 {
		globalPath = globalPath[:len(globalPath)-1]
	}
}

func reinitialiseGlobalPath() {
	globalPath = make([]string, 0)
}