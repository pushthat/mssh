package main

import (
	"github.com/manifoldco/promptui"
	"regexp"
	"errors"
	"sort"
	"os"
	"fmt"
)

func clearTheScreen() {
	fmt.Printf("\033[0;0H")
}

func askUserUser() string {
	validate := func(input string) error {
		if len(input) > 0 {
			return nil
		} else {
			return errors.New("User must contain at least one character")
		}
	}
	prompt := promptui.Prompt{
		Label: "User",
		Validate: validate,
	}
	result, err := prompt.Run()
	check(err)
	return result
}

func askUserIp() string {
	validate := func(input string) error {
		includeRegex, err := regexp.Compile(`^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)
		check(err)
		if includeRegex.Match([]byte(input)) {
			return nil
		} else {
			return errors.New("Invalid Ip")
		}
	}

	prompt := promptui.Prompt{
		Label:    "Ip",
		Validate: validate,
	}
	result, err := prompt.Run()
	check(err)
	return result
}

func askUserPassword() string {
	prompt := promptui.Prompt{
		Label: "Password",
		Mask:  '*',
	}
	result, err := prompt.Run()
	check(err)
	return result
}

func askUserSshKeyPasscode(pathSshKey *string) string {
	validate := func(input string) error {
		err := isPublicKeyPasswordCorrect(pathSshKey, &input)
		if err == nil {
			return nil
		} else {
			return err
		}
	}

	prompt := promptui.Prompt{
		Label:    "Password of the ssh key",
		Mask:     '*',
		Validate: validate,
	}
	result, err := prompt.Run()
	check(err)
	return result
}

func askUserSshPath() (string, string) {
	sshPassCode := ""

	validate := func(input string) error {
		if input == "" {
			return nil
		}

		passNeeded, err := doesPublicKeyNeedPassword(&input)
		if err == nil && passNeeded {
			return nil
		}

		err = isPublicKeyFileFormatted(&input)
		if err == nil {
			return nil
		} else {
			return errors.New(err.Error())
		}
	}

	prompt := promptui.Prompt{
		Label:    "Ssh key path (skip if none)",
		Validate: validate,
		Default:  os.Getenv("HOME") + "/.ssh/id_rsa",
	}
	result, err := prompt.Run()
	check(err)
	pass, _ := doesPublicKeyNeedPassword(&result)
	if pass {
		sshPassCode = askUserSshKeyPasscode(&result)
	}
	return result, sshPassCode
}

func askUserName() string {
	validate := func(input string) error {
		if len(input) > 0 {
			return nil
		} else {
			return errors.New("Name must contain at least one character")
		}
	}

	prompt := promptui.Prompt{
		Label:    "Name",
		Validate: validate,
	}
	result, err := prompt.Run()
	check(err)
	return result
}

func askUserHost() {
	ip := askUserIp()
	user := askUserUser()
	sshKeyPath, sshPassword := askUserSshPath()
	pass := ""
	if sshKeyPath == "" {
		pass = askUserPassword()
	}
	name := askUserName()

	host := Host{ip, user, sshKeyPath, true}
	if canHostBeReached(&host, &pass, &sshPassword) {
		createHost(&host, name)
	} else {
		clearTheScreen()
		fmt.Println("This host cannot be reached...")
		askUserHost()
	}
}

func askUserProject() {
	name := askUserName()

	createProject(name)
}

func convertKeyToReadable(keys *map[string]interface{}) []Readable {
	sliceReadable := make([]Readable, 0)

	for key, element := range *keys {
		_, val := element.(map[string]interface{})["IsHost"]
		if val {
			readable := Readable{key, "Host"}
			sliceReadable = append(sliceReadable, readable)
		} else {
			readable := Readable{key, "Repository"}
			sliceReadable = append(sliceReadable, readable)
		}
	}
	return sliceReadable
}

func addOption(keys []Readable) []Readable {
	keys = append(keys, Readable{addHostKey, "Option"})
	keys = append(keys, Readable{addProjectKey, "Option"})
	if len(globalPath) > 0 {
		keys = append(keys, Readable{goBackKey, "Option"})
	}
	return keys
}

func addRmOption(keys []Readable) []Readable {
	keys = append(keys, Readable{rmTheRepoKey, "Option"})
	if len(globalPath) > 0 {
		keys = append(keys, Readable{goBackKey, "Option"})
	}
	return keys
}

func orderReadableSliceAlphabetically(keys []Readable) []Readable {
	sort.Slice(keys, func(i, j int) bool { return keys[i].Name < keys[j].Name })
	return keys
}

func askToUser(keys map[string]interface{}, label string) string {
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   "{{ .Name | red }} ({{ .Desc | red }})",
		Inactive: "{{ .Name }} ({{ .Desc }})",
		Selected: "{{ .Name }}",
		Details: `
--------- DESC ----------
{{ "Name:" | faint }}	{{ .Name }}
{{ "Type:" | faint }}	{{ .Desc }}`,
	}
	readables := convertKeyToReadable(&keys)
	readables = orderReadableSliceAlphabetically(readables)
	readables = addOption(readables)
	prompt := promptui.Select{
		Label:     label,
		Items:     readables,
		Templates: templates,
		Size:      10,
	}

	i, _, err := prompt.Run()

	result := readables[i].Name

	check(err)
	if result == addHostKey {
		askUserHost()
	}
	if result == addProjectKey {
		askUserProject()
	}
	return result
}

func isUserIsSure(valToDel string) bool {
	readables := make([]Readable, 2)
	readables[0] = Readable{"Yes", "option"}
	readables[1] = Readable{"No", "option"}
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   "{{ .Name | red }} ({{ .Desc | red }})",
		Inactive: "{{ .Name }} ({{ .Desc }})",
		Selected: "{{ .Name }}",
		Details: `
--------- DESC ----------
{{ "Name:" | faint }}	{{ .Name }}
{{ "Type:" | faint }}	{{ .Desc }}`,
	}
	prompt := promptui.Select{
		Label:     "Are you sure you want to delete " + valToDel + " ?",
		Items:     readables,
		Templates: templates,
		Size:      10,
	}
	i, _, err := prompt.Run()
	result := readables[i].Name
	check(err)
	return result == "Yes"
}

func askToUserHostToRm(keys map[string]interface{}, label string) string {
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   "{{ .Name | red }} ({{ .Desc | red }})",
		Inactive: "{{ .Name }} ({{ .Desc }})",
		Selected: "{{ .Name }}",
		Details: `
--------- DESC ----------
{{ "Name:" | faint }}	{{ .Name }}
{{ "Type:" | faint }}	{{ .Desc }}`,
	}
	readables := convertKeyToReadable(&keys)
	readables = orderReadableSliceAlphabetically(readables)
	if len(globalPath) > 0 {
		fmt.Println(globalPath)
		readables = addRmOption(readables)
	}
	prompt := promptui.Select{
		Label:     label,
		Items:     readables,
		Templates: templates,
		Size:      10,
	}

	i, _, err := prompt.Run()

	result := readables[i].Name

	check(err)
	if result == rmTheRepoKey {
		if isUserIsSure(result) {
			return result
		} else {
			askToUserHostToRm(keys, label)
		}
	}
	return result
}
