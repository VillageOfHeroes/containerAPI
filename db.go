package main

import (
	"encoding/json"
	"errors"
	"log"
	"os"
)

type DB struct {
	LastPort     int                  `json:"port"`
	PortsList    map[int]bool         `json:"ports_list"`
	ContainerMap map[string]Container `json:"container_map"`
}

type Container struct {
	Image string `json:"image"`
	Name  string `json:"name"`
	Up    bool   `json:"online"`
	Port  []int  `json:"port"`
	Id    string `json:"id"`
}

var db DB

func InitDB() {
	db.LastPort = 25595
	db.PortsList = make(map[int]bool)
	db.ContainerMap = make(map[string]Container)
}

func GetContainer(name string) (Container, bool) {
	container, ok := db.ContainerMap[name]
	return container, ok
}

func SetContainerUpState(container Container, state bool) {
	c, ok := db.ContainerMap[container.Name]
	if ok {
		c.Up = state
		db.ContainerMap[container.Name] = c
	}

}

func AddContainer(container Container) {
	db.ContainerMap[container.Name] = container
}

func RemoveContainer(container Container) {
	delete(db.ContainerMap, container.Name)
	json.Marshal(true)
}

func GetPort() (int, error) {
	startport := db.LastPort
	for {
		db.LastPort += 10

		if db.LastPort > 27000 {
			db.LastPort = 25595
			log.Printf("Port: %d\n", db.LastPort)
		}

		elem, ok := db.PortsList[db.LastPort]
		if !ok || ok && !elem {
			break
		}

		if db.LastPort == startport {
			return -1, errors.New("no Port availble")
		}
	}
	db.PortsList[db.LastPort] = true
	return db.LastPort, nil
}

func FreePort(port int) {
	db.PortsList[port] = false
}

func Save() {
	file, err := os.OpenFile("save.json", os.O_TRUNC|os.O_WRONLY, os.ModePerm)

	if err == nil {
		log.Print("Saving now...")
		err = json.NewEncoder(file).Encode(db)
		if err != nil {
			log.Print(err.Error())
		}
	} else {
		log.Fatal("COULD NOT SAVE CONTAINER!")
	}

	defer file.Close()

}

func Load() {
	file, err := os.OpenFile("save.json", os.O_RDONLY, os.ModePerm)

	if err == nil {
		err = json.NewDecoder(file).Decode(&db)
		if err != nil {
			log.Print(err.Error())
		}
	} else {
		log.Fatal("COULD NOT LOAD CONTAINER!")
	}

	LoadImages()

	defer file.Close()
}
