package main

import (
	"encoding/json"
	"errors"
	"log"
)

type DB struct {
	lastPort     int
	portsList    map[int]bool
	containerMap map[string]Container
}

type Container struct {
	Image string `json:"image"`
	Name  string `json:"name"`
	Up    bool   `json:"online"`
	Port  int    `json:"port"`
	Id    string `json:"id"`
}

var db DB

func InitDB() {
	db.lastPort = 25595
	db.portsList = make(map[int]bool)
	db.containerMap = make(map[string]Container)
}

func GetContainer(name string) (Container, bool) {
	container, ok := db.containerMap[name]
	return container, ok
}

func SetContainerUpState(container Container, state bool) {
	c, ok := db.containerMap[container.Name]
	if ok {
		c.Up = state
		db.containerMap[container.Name] = c
	}

}

func AddContainer(container Container) {
	db.containerMap[container.Name] = container
}

func RemoveContainer(container Container) {
	delete(db.containerMap, container.Name)
	json.Marshal(true)
}

func GetPort() (int, error) {
	startport := db.lastPort
	for {
		db.lastPort += 10

		if db.lastPort > 27000 {
			db.lastPort = 25595
			log.Printf("Port: %d\n", db.lastPort)
		}

		elem, ok := db.portsList[db.lastPort]
		if !ok || ok && !elem {
			break
		}

		if db.lastPort == startport {
			return -1, errors.New("no Port availble")
		}
	}
	db.portsList[db.lastPort] = true
	return db.lastPort, nil
}

func FreePort(port int) {
	db.portsList[port] = false
}
