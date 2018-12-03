package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/only1isus/majorProj/util"
)

const (
	host     string = "localhost"
	grpcPort string = ":8081"
	port     string = ":8080"
)

// Userinfo defines the user data structure
// type user util.User

func postUser(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Println(err)
	}

	var u util.User
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
	}

	if err := json.Unmarshal(body, &u); err != nil {
		fmt.Println(err)
	}
	err = u.Add()
	if err != nil {
		w.WriteHeader(500)
		fmt.Println(err)
	} else {
		w.Header()
		w.WriteHeader(200)
	}
}

// returns the sensor data by type
func getSensorData(w http.ResponseWriter, r *http.Request) {

	device := r.URL.Path[len("/api/sensor/"):]
	device = strings.ToLower(device)
	data, err := util.GetSensorDataByType(device)
	d, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("got an error while trying encode response")
		http.Error(w, err.Error(), 500)

	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(d)
	}

}

// get all the logs
func getLogs(w http.ResponseWriter, r *http.Request) {
	logs, err := util.GetLogs()
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), 500)
	}
	l, err := json.Marshal(logs)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), 500)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(l)
	}
}

// changeSettings edits the config file of the system
func changeSettings(w http.ResponseWriter, r *http.Request) {
	fmt.Println("trying to change settings")
	if err := r.ParseForm(); err != nil {
		fmt.Println(err)
	}

	var u util.User
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
	}

	if err := json.Unmarshal(body, &u); err != nil {
		fmt.Println(err)
	}
	err = u.Add()
	if err != nil {
		w.WriteHeader(500)
		fmt.Println(err)
	} else {
		w.Header()
		w.WriteHeader(200)
	}

}

func main() {
	allowedHeaders := handlers.AllowedHeaders([]string{"application/json", "application/x-www-form-urlencoded", "multipart/form-data"})
	allowedOrigins := handlers.AllowedOrigins([]string{"*"})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	router := mux.NewRouter()
	log.Printf("server running pn port %s...", port)
	router.HandleFunc("/newuser", postUser).Methods("POST")
	router.HandleFunc("/api/sensor/{type}", getSensorData).Methods("GET")
	router.HandleFunc("/api/logs", getLogs).Methods("GET")
	router.HandleFunc("/api/settings", changeSettings).Methods("POST")
	http.ListenAndServe(port, handlers.CORS(allowedHeaders, allowedOrigins, allowedMethods)(router))

}
