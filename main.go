package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

var port int
var portsList map[int]bool

func main() {
	portsList = make(map[int]bool)
	port = 25595
	r := mux.NewRouter()
	r.HandleFunc("/getPort", GetPortHandler)
	r.HandleFunc("/freePort/{port}", FreePortHandler)
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
func FreePortHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	elem, ok := vars["port"]
	elem2, err := strconv.Atoi(elem)
	if ok && err == nil {
		w.WriteHeader(http.StatusOK)
		portsList[elem2] = false
		json.NewEncoder(w).Encode(map[string]string{"ok": "Worked"})
	} else {
		w.WriteHeader(http.StatusExpectationFailed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Parse Error of Port. Port doesnt seem to be valid"})
	}

}

func GetPortHandler(w http.ResponseWriter, r *http.Request) {
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
			json.NewEncoder(w).Encode(map[string]string{"error": "No Port available"})
			return
		}
	}
	portsList[port] = true
	json.NewEncoder(w).Encode(map[string]int{"port": port})

}
