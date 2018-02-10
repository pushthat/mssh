package main

import (
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"errors"
	"strings"
	"encoding/pem"
	"time"
	"os"
	"os/exec"
)

func doesPublicKeyNeedPassword(filePath *string) (bool, error) {
	if !fileExist(filePath) {
		return false, errors.New("Invalid file path")
	}

	buffer, err := ioutil.ReadFile(*filePath)
	if err != nil {
		return false, err
	}

	block, _ := pem.Decode(buffer)
	return strings.Contains(block.Headers["Proc-Type"], "ENCRYPTED"), nil
}

func isPublicKeyPasswordCorrect(filePath *string, password *string) error {
	if !fileExist(filePath) {
		return errors.New("Invalid file path")
	}
	buffer, err := ioutil.ReadFile(*filePath)
	if err != nil {
		return err
	}

	_, err = ssh.ParsePrivateKeyWithPassphrase(buffer, []byte(*password))
	if err != nil {
		return err
	}
	return nil
}

func isPublicKeyFileFormatted(filePath *string) error {
	if !fileExist(filePath) {
		return errors.New("Invalid file path")
	}
	buffer, err := ioutil.ReadFile(*filePath)
	if err != nil {
		return err
	}

	_, err = ssh.ParsePrivateKey(buffer)
	if err != nil {
		return err
	}
	return nil
}

func publicKeyFile(file *string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(*file)
	if err != nil {
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}

func publicKeyWithPassword(file *string, password *string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(*file)
	if err != nil {
		return nil
	}

	key, err := ssh.ParsePrivateKeyWithPassphrase(buffer, []byte(*password))
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}

func configWithSshKey(user *string, sshKeyPath *string) *ssh.ClientConfig {
	config := &ssh.ClientConfig{
		User: *user,
		Auth: []ssh.AuthMethod{
			publicKeyFile(sshKeyPath),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	return config
}
func configWithShhKeyWithPassword(user *string, sshKeyPath *string, password *string) *ssh.ClientConfig {
	config := &ssh.ClientConfig{
		User: *user,
		Auth: []ssh.AuthMethod{
			publicKeyWithPassword(sshKeyPath, password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	return config
}

func configWithPassword(user *string, password *string) *ssh.ClientConfig {
	config := &ssh.ClientConfig{
		User: *user,
		Auth: []ssh.AuthMethod{
			ssh.Password(*password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	return config
}

func findCorrectShhConfigForThisHost(host *Host, pass *string, sshPass *string) *ssh.ClientConfig {
	if len(host.SshKeyPath) > 0 {
		if len(*sshPass) > 0 {
			return configWithShhKeyWithPassword(&host.User, &host.SshKeyPath, sshPass)
		} else {
			return configWithSshKey(&host.User, &host.SshKeyPath)
		}
	} else {
		return configWithPassword(&host.User, sshPass)
	}
}

func initiateConnection(host *Host, result chan error, pass *string, sshPass *string) {
	sshConfig := findCorrectShhConfigForThisHost(host, pass, sshPass)

	connection, err := ssh.Dial("tcp", host.Ip+":22", sshConfig)
	if err != nil {
		result <- err
	}
	_, err = connection.NewSession()
	if err != nil {
		result <- err
	}
	result <- nil
}

func lauchSshClientBet(host *Host) {
	cmd := exec.Command("ssh",  host.User + "@" + host.Ip)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func canHostBeReached(host *Host, pass *string, sshPass *string) bool {
	result := make(chan error)

	go initiateConnection(host, result, pass, sshPass)

	select {
	case res := <-result:
		if res == nil {
			return true
		} else {
			return false
		}
	case <-time.After(time.Second * 1):
		return false
	}

}
