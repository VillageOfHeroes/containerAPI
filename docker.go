package main

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func DeleteContainer(container Container) error {
	if container.Up {
		return errors.New("Container is still online")
	}
	RemoveContainer(container)
	FreePort(container.Port)

	return nil
}

func CreateContainer(image string) (Container, error) {

	empty := Container{}
	if image == "itzg/minecraft-server" {
		port, err := GetPort()
		if err != nil {
			return empty, err
		}
		portStr := strconv.Itoa(port)
		cmd := exec.Command("docker", "run", "-d", "-e", "EULA=TRUE", "--cpus=2", "-p", portStr+":25565", "itzg/minecraft-server")
		log.Print(cmd.String())
		cmd.Stderr = os.Stderr

		out, err2 := cmd.Output()
		id := string(out)

		if err2 != nil {
			log.Fatal(err2)
			FreePort(port)
			return empty, err2
		}

		cmd = exec.Command("sh", "./getName.sh", id)
		cmd.Stderr = os.Stderr

		out, err = cmd.Output()
		name := string(out)
		name = strings.Trim(name, "\n ")

		if err != nil {
			log.Fatal(err)
			FreePort(port)
			return empty, err
		}

		container := Container{Image: image, Port: port, Name: name, Id: id, Up: true}

		AddContainer(container)
		return container, nil
	} else {
		return empty, errors.New("image currently not supported")
	}

}

func StartContainer(container Container) error {
	cmd := exec.Command("docker", "start", container.Name)
	log.Print(cmd.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	SetContainerUpState(container, true)
	return err
}

func StopContainer(container Container) error {
	cmd := exec.Command("docker", "stop", container.Name)
	log.Print(cmd.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	SetContainerUpState(container, false)
	return err
}

func SendCommand(container Container, command string) (string, error) {
	cmd := exec.Command("docker", "exec", container.Name, "rcon-cli", command)
	log.Print(cmd.String())
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	return string(out), err
}
