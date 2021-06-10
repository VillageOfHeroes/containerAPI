package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

var imageDB map[string]Image = make(map[string]Image)

type Image struct {
	Name         string            `json:"name"`
	Ports        []Port            `json:"ports"`
	Image        string            `json:"image_name"`
	Environments map[string]string `json:"environment_vars"`
	CommandArgs  []string          `json:"command_args"`
}

type Port struct {
	Port int  `json:"port"`
	Udp  bool `json:"udp"`
}

func CreateImage(name string, ports []Port, image_name string, envs map[string]string) {
	image := Image{Name: name, Ports: ports, Image: image_name, Environments: envs}
	imageDB[name] = image
	SaveImage(image)
}

func LoadImage(name string) (Image, error) {
	log.Print("Loading now Image: " + name)
	file, err := os.OpenFile("configs/images/"+name+".json", os.O_RDONLY, os.ModePerm)

	image := Image{}

	if err == nil {

		err = json.NewDecoder(file).Decode(&image)
		defer file.Close()
		if err != nil {
			log.Print(err.Error())
			return image, errors.New("COULD NOT Load IMAGE!")
		}
		imageDB[name] = image
		log.Print("Loading done Image: " + name)
		return image, nil
	} else {
		return image, errors.New("COULD NOT Load IMAGE!")
	}
}

func SaveImage(image Image) {
	log.Print("Saving now Image: " + image.Name)
	file, err := os.OpenFile("configs/images/"+image.Name+".json", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)

	if err == nil {

		err = json.NewEncoder(file).Encode(image)
		if err != nil {
			log.Print(err.Error())
		} else {
			log.Print("Saving done Image: " + image.Name)
		}
	} else {
		log.Fatal("COULD NOT SAVE IMAGE!")
	}

	defer file.Close()

}

func LoadImages() {
	//LoadImage("minecraft-server")
	//LoadImage("factorio-server")

	files, err := ioutil.ReadDir("configs/images/")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		_, err = LoadImage(strings.TrimSuffix(file.Name(), path.Ext(file.Name())))
		if err != nil {
			log.Fatal(err.Error())
		}
	}

}
