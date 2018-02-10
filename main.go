package main

import (
	"github.com/docopt/docopt-go"
	"fmt"
	"os"
)

var msshVersion = "0.5"
var addHostKey = "+ Host"
var addProjectKey = "+ Project"
var goBackKey = "<-Back"
var rmTheRepoKey = "Delete the project"
var pathToStorageFile = os.Getenv("HOME") + "/.mssh_storage.json"
var sshMap map[string]interface{}
var globalPath = make([]string, 0)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func getCurrentFromGivenPath(origin *map[string]interface{}, givenPath []string) map[string]interface{} {
	current := *origin

	for _, element := range givenPath {
		tmp_val := current[element]
		current = tmp_val.(map[string]interface{})
	}
	return current
}

func getCurrentFromGlobalPath(origin *map[string]interface{}) map[string]interface{} {
	current := *origin

	for _, element := range globalPath {
		tmpVal := current[element]
		fmt.Println(element)
		current = tmpVal.(map[string]interface{})
	}
	return current
}

func createHost(host *Host, name string) {
	current := getMapFromFile()

	node := getCurrentFromGlobalPath(&current)
	node[name] = *host
	updateConfigFile(current)
	globalPath = make([]string, 0)
}

func remove(name string) {
	current := getMapFromFile()

	givenPath := globalPath[:len(globalPath)-1]
	node := getCurrentFromGivenPath(&current, givenPath)
	delete(node, name)
	updateConfigFile(current)
}

func createProject(name string) {
	current := getMapFromFile()

	node := getCurrentFromGlobalPath(&current)
	node[name] = make(map[string]interface{})
	updateConfigFile(current)
}

func getArgument() map[string]interface{} {
	usage := `mssh.

Usage:
  mssh co
  mssh rm
  mssh --version

Options:
  -h --help     Show this screen.`

	arguments, _ := docopt.Parse(usage, nil, true, msshVersion, false)
	return arguments
}

func getKeysOutOfThatMap(mapv map[string]interface{}) []string {
	keys := make([]string, len(mapv))
	i := 0
	for k := range mapv {
		keys[i] = k
		i++
	}
	return keys
}

func getHostFromUser(current interface{}) (Host) {
	_, val := current.(map[string]interface{})["IsHost"]
	clearTheScreen()
	if val {
		return createFromMap(current.(map[string]interface{}))
	}
	userResponse := askToUser(current.(map[string]interface{}), "Select repo or host")
	if userResponse == addHostKey {
		return getHostFromUser(getMapFromFile())
	} else if userResponse == addProjectKey {
		reinitialiseGlobalPath()
		return getHostFromUser(getMapFromFile())
	} else if userResponse == goBackKey {
		changeToUpperLevelProject()
		return getHostFromUser(getMapAtGlobalPath())
	} else {
		updateGlobalPath(userResponse)
	}
	return getHostFromUser(current.(map[string]interface{})[userResponse])
}

func getHostToDelFromUser(current interface{}) (bool) {
	_, val := current.(map[string]interface{})["IsHost"]
	clearTheScreen()
	if val {
		if isUserIsSure(createFromMap(current.(map[string]interface{})).Ip) {
			return true
		}
		return false
	}
	userResponse := askToUserHostToRm(current.(map[string]interface{}), "Select repo or host")
	if userResponse == rmTheRepoKey {
		return true
	} else if userResponse == goBackKey {
		changeToUpperLevelProject()
		return getHostToDelFromUser(getMapAtGlobalPath())
	} else {
		updateGlobalPath(userResponse)
	}
	return getHostToDelFromUser(current.(map[string]interface{})[userResponse])
}

func main() {
	args := getArgument()

	//Verify if the file exist
	if !fileExist(&pathToStorageFile) {
		createStorageFile()
	}

	if args["--version"].(bool) {
		fmt.Println(msshVersion)
	} else if args["rm"].(bool) {
		current := getMapFromFile()
		if getHostToDelFromUser(current) {
			remove(globalPath[len(globalPath)-1])
 		} else {
			main()
		}
	} else if args["co"].(bool) {
		////Get the info from the file
		current := getMapFromFile()
		host := getHostFromUser(current)

		fmt.Println("You are connected to : " + host.Ip)
		lauchSshClientBet(&host)
	}
}
