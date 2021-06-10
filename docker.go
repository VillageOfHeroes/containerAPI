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
	for i := range container.Port {
		FreePort(i)
	}

	cmd := exec.Command("docker", "rm", "-v", container.Name)
	log.Print(cmd.String())
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()

	return err
}

func CreateContainer(image_name string) (Container, error) {

	empty := Container{}
	image, ok := imageDB[image_name]
	if !ok {
		return empty, errors.New("image currently not supported")
	}

	ports := make([]int, len(image.Ports))
	for i := 0; i < len(image.Ports); i++ {
		port, err := GetPort()
		if err != nil {
			return empty, err
		}
		ports[i] = port
	}
	// , "-e", "EULA=TRUE", "--cpus=2", "-p", portStr + ":25565", "itzg/minecraft-server"

	args := []string{"run", "-d", "--cpus=2"}

	for i := 0; i < len(image.Ports); i++ {
		if image.Ports[i].Udp {
			args = append(args, "-p", strconv.Itoa(ports[i])+":"+strconv.Itoa(image.Ports[i].Port)+"/udp")
		}
		args = append(args, "-p", strconv.Itoa(ports[i])+":"+strconv.Itoa(image.Ports[i].Port)+"/tcp")
	}

	for k, v := range image.Environments {
		args = append(args, "-e", k+"="+v)
	}

	args = append(args, image.Image)

	for i := 0; i < len(image.CommandArgs); i++ {
		args = append(args, image.CommandArgs[i])
	}

	cmd := exec.Command("docker", args[0:]...)
	log.Print(cmd.String())
	cmd.Stderr = os.Stderr

	out, err2 := cmd.Output()
	id := strings.Trim(string(out), " \n")

	if err2 != nil {
		log.Fatal(err2)
		//FreePort(port)
		return empty, err2
	}

	cmd = exec.Command("sh", "./getName.sh", id)
	cmd.Stderr = os.Stderr

	out, err2 = cmd.Output()
	name := string(out)
	name = strings.Trim(name, "\n ")

	if err2 != nil {
		log.Fatal(err2)
		//FreePort(port)
		return empty, err2
	}

	container := Container{Image: image_name, Port: ports, Name: name, Id: id, Up: true}

	AddContainer(container)
	return container, nil

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
