package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// XXXXXX main XXXXX

func main() {

	InitDB()
	Load()

	r := mux.NewRouter()
	r.HandleFunc("/getPort/{username}/{token}", GetPortHandler)
	r.HandleFunc("/freePort/{port}/{username}/{token}", FreePortHandler)
	r.HandleFunc("/createServer/{image_name}/{username}/{token}", CreateServerHandler)
	r.HandleFunc("/deleteServer/{name}/{username}/{token}", DeleteServerHandler)
	r.HandleFunc("/startServer/{name}/{username}/{token}", StartServerHandler)
	r.HandleFunc("/stopServer/{name}/{username}/{token}", StopServerHandler)
	r.HandleFunc("/sendCommand/{name}/{command}/{username}/{token}", SendCommandHandler)

	// needs admin

	r.HandleFunc("/secretInfos/{username}/{token}", SecretServerHandler)

	// without login

	r.HandleFunc("/getServerInfo/{name}", GetServerInfoHandler)
	r.HandleFunc("/getServerInfos/", GetServerInfosHandler)
	r.HandleFunc("/getImages/", GetImagesHandler)
	r.HandleFunc("/getToken/{username}", GetTokenHandler)
	http.Handle("/", r)

	srv := &http.Server{
		Handler: r,
		Addr:    "127.0.0.1:8001",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			Save()
			log.Fatal(err)
		}
	}()

	// Setting up signal capturing
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Waiting for SIGINT (pkill -2)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}

// XXXXX Functions XXXXX

// XXXXX Handler XXXXX

func GetTokenHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	username, ok := vars["username"]

	user, ok2 := GetUser(username)
	if ok && ok2 {
		token := genToken(user)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"token": token})
	} else {
		w.WriteHeader(http.StatusExpectationFailed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Could not find User"})
	}
}

func GetServerInfoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	name, ok := vars["name"]
	if !ok {
		w.WriteHeader(http.StatusExpectationFailed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Please provide Name"})
		return
	}
	container, ok2 := GetContainer(name)
	if !ok2 {
		w.WriteHeader(http.StatusExpectationFailed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Container does not exist"})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(container)
	log.Print(container)
}

func GetServerInfosHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(db.ContainerMap)
	log.Print(db)
}

func GetImagesHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	images := []string{}
	for k, _ := range imageDB {
		images = append(images, k)
	}

	json.NewEncoder(w).Encode(images)
	log.Print(db)
}

// HANDLER WITH TOKENS

func FreePortHandler(w http.ResponseWriter, r *http.Request) {
	err := CheckPermission(r, ADMIN)
	if err == nil {
		vars := mux.Vars(r)

		elem, ok := vars["port"]
		elem2, err := strconv.Atoi(elem)
		if ok && err == nil {
			w.WriteHeader(http.StatusOK)
			FreePort(elem2)
			json.NewEncoder(w).Encode(map[string]string{"ok": "Worked"})
		} else {
			w.WriteHeader(http.StatusExpectationFailed)
			json.NewEncoder(w).Encode(map[string]string{"error": "Parse Error of Port. Port doesnt seem to be valid"})
		}
	} else {
		w.WriteHeader(http.StatusExpectationFailed)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
	}
}

func GetPortHandler(w http.ResponseWriter, r *http.Request) {
	err := CheckPermission(r, ADMIN)
	if err == nil {
		port, err := GetPort()
		if err != nil {
			w.WriteHeader(http.StatusExpectationFailed)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		} else {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]int{"port": port})
		}
	} else {
		w.WriteHeader(http.StatusExpectationFailed)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
	}
}

func CreateServerHandler(w http.ResponseWriter, r *http.Request) {
	err := CheckPermission(r, ADMIN)
	if err == nil {
		vars := mux.Vars(r)

		image_name, ok := vars["image_name"]
		if !ok {
			w.WriteHeader(http.StatusExpectationFailed)
			json.NewEncoder(w).Encode(map[string]string{"error": "Please provide An Image"})
			return
		}

		container, err := CreateContainer(image_name)
		if err != nil {
			w.WriteHeader(http.StatusExpectationFailed)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"port": container.Port, "name": container.Name})
	} else {
		w.WriteHeader(http.StatusExpectationFailed)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
	}

}

func DeleteServerHandler(w http.ResponseWriter, r *http.Request) {
	err := CheckPermission(r, ADMIN)
	if err == nil {
		vars := mux.Vars(r)

		name, ok := vars["name"]
		if !ok {
			w.WriteHeader(http.StatusExpectationFailed)
			json.NewEncoder(w).Encode(map[string]string{"error": "Please provide Name"})
			return
		}
		container, ok2 := GetContainer(name)
		if !ok2 {
			w.WriteHeader(http.StatusExpectationFailed)
			json.NewEncoder(w).Encode(map[string]string{"error": "Container does not exist"})
			return
		}

		err := DeleteContainer(container)
		if err == nil {
			w.WriteHeader(http.StatusOK)

			json.NewEncoder(w).Encode(map[string]string{"ok": "Server Deleted Successfully"})
		} else {
			w.WriteHeader(http.StatusExpectationFailed)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		}
	} else {
		w.WriteHeader(http.StatusExpectationFailed)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
	}
}

func StartServerHandler(w http.ResponseWriter, r *http.Request) {
	err := CheckPermission(r, ADMIN)
	if err == nil {
		vars := mux.Vars(r)

		name, ok := vars["name"]
		if !ok {
			w.WriteHeader(http.StatusExpectationFailed)
			json.NewEncoder(w).Encode(map[string]string{"error": "Please provide Name"})
			return
		}
		container, ok2 := GetContainer(name)
		if !ok2 {
			w.WriteHeader(http.StatusExpectationFailed)
			json.NewEncoder(w).Encode(map[string]string{"error": "Container does not exist"})
			return
		}
		err := StartContainer(container)
		if err == nil {
			w.WriteHeader(http.StatusOK)

			json.NewEncoder(w).Encode(map[string]string{"ok": "Server Started Successfully"})
		} else {
			w.WriteHeader(http.StatusExpectationFailed)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		}
	} else {
		w.WriteHeader(http.StatusExpectationFailed)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
	}
}

func StopServerHandler(w http.ResponseWriter, r *http.Request) {
	err := CheckPermission(r, ADMIN)
	if err == nil {
		vars := mux.Vars(r)

		name, ok := vars["name"]
		if !ok {
			w.WriteHeader(http.StatusExpectationFailed)
			json.NewEncoder(w).Encode(map[string]string{"error": "Please provide Name"})
			return
		}
		container, ok2 := GetContainer(name)
		if !ok2 {
			w.WriteHeader(http.StatusExpectationFailed)
			json.NewEncoder(w).Encode(map[string]string{"error": "Container does not exist"})
			return
		}
		err := StopContainer(container)
		if err == nil {
			w.WriteHeader(http.StatusOK)

			json.NewEncoder(w).Encode(map[string]string{"ok": "Server Stopped Successfully"})
		} else {
			w.WriteHeader(http.StatusExpectationFailed)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		}
	} else {
		w.WriteHeader(http.StatusExpectationFailed)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
	}
}

func SendCommandHandler(w http.ResponseWriter, r *http.Request) {
	err := CheckPermission(r, ADMIN)
	if err == nil {
		vars := mux.Vars(r)

		name, ok := vars["name"]
		command, ok2 := vars["command"]
		if !ok {
			w.WriteHeader(http.StatusExpectationFailed)
			json.NewEncoder(w).Encode(map[string]string{"error": "Please provide Name"})
			return
		}
		if !ok2 {
			w.WriteHeader(http.StatusExpectationFailed)
			json.NewEncoder(w).Encode(map[string]string{"error": "Please provide A Command"})
			return
		}
		container, ok2 := GetContainer(name)
		if !ok2 {
			w.WriteHeader(http.StatusExpectationFailed)
			json.NewEncoder(w).Encode(map[string]string{"error": "Container does not exist"})
			return
		}
		output, err := SendCommand(container, command)
		if err == nil {
			w.WriteHeader(http.StatusOK)

			json.NewEncoder(w).Encode(map[string]string{"out": output})
		} else {
			w.WriteHeader(http.StatusExpectationFailed)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		}
	} else {
		w.WriteHeader(http.StatusExpectationFailed)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
	}
}

func SecretServerHandler(w http.ResponseWriter, r *http.Request) {
	err := CheckPermission(r, ADMIN)
	if err == nil {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(db.ContainerMap)
		log.Print(db)

		json.NewEncoder(w).Encode(imageDB)
		log.Print(imageDB)
	} else {
		w.WriteHeader(http.StatusExpectationFailed)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
	}
}
