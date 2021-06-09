package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

var port int
var portsList map[int]bool

// XXXXXX main XXXXX

func main() {
	portsList = make(map[int]bool)
	port = 25595
	r := mux.NewRouter()
	r.HandleFunc("/getPort", GetPortHandler)
	r.HandleFunc("/freePort/{port}", FreePortHandler)
	r.HandleFunc("/createServer/{image}", CreateServerHandler)
	http.Handle("/", r)

	srv := &http.Server{
		Handler: r,
		Addr:    "127.0.0.1:8000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

// XXXXX Functions XXXXX

func getPort() (int, error) {
	startport := port
	for {
		port += 10

		if port > 27000 {
			port = 25595
			log.Printf("Port: %d\n", port)
		}

		elem, ok := portsList[port]
		if !ok || ok && !elem {
			break
		}

		if port == startport {
			return -1, errors.New("No Port availble")
		}
	}
	portsList[port] = true
	return port, nil
}

func freePort(port int) {
	portsList[port] = false
}

func createServer(image string) (int, error) {

	if image == "itzg/minecraft-server" {
		port, err := getPort()
		if err != nil {
			return port, err
		}
		portStr := strconv.Itoa(port)
		cmd := exec.Command("docker", "run", "-d", "-e", "EULA=TRUE", "--cpus=2", "-p", portStr+":25565", "itzg/minecraft-server")
		log.Print(cmd.String())
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()

		if err != nil {
			log.Fatal(err)
			freePort(port)
			return port, err
		}

		return port, nil
	} else {
		return port, errors.New("Image currently not supported")
	}

}

// XXXXX Handler XXXXX

func FreePortHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	elem, ok := vars["port"]
	elem2, err := strconv.Atoi(elem)
	if ok && err == nil {
		w.WriteHeader(http.StatusOK)
		freePort(elem2)
		json.NewEncoder(w).Encode(map[string]string{"ok": "Worked"})
	} else {
		w.WriteHeader(http.StatusExpectationFailed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Parse Error of Port. Port doesnt seem to be valid"})
	}

}

func GetPortHandler(w http.ResponseWriter, r *http.Request) {
	port, err := getPort()
	if err != nil {
		w.WriteHeader(http.StatusExpectationFailed)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
	} else {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]int{"port": port})
	}

}

func CreateServerHandler(w http.ResponseWriter, r *http.Request) {
	port, err := createServer("itzg/minecraft-server")
	if err != nil {
		w.WriteHeader(http.StatusExpectationFailed)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int{"port": port})

}
