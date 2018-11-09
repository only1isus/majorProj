package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

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
	u.CreatedAt = time.Now().Unix()
	if err := util.CreateUser(&u); err != nil {
		w.Header().Set("message", "user already exists")
		w.WriteHeader(500)
		fmt.Println(err)
	} else {
		w.Header()
		w.WriteHeader(200)
	}
}

func getSensorData(w http.ResponseWriter, r *http.Request) {

	device := r.URL.Path[len("/api/sensor/"):]
	device = strings.ToLower(device)
	data, err := util.GetSensorData(device)
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

func main() {
	allowedHeaders := handlers.AllowedHeaders([]string{"application/json", "application/x-www-form-urlencoded", "multipart/form-data"})
	allowedOrigins := handlers.AllowedOrigins([]string{"*"})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	router := mux.NewRouter()
	log.Printf("server running pn port %s...", port)
	router.HandleFunc("/newuser", postUser).Methods("POST")
	router.HandleFunc("/api/sensor/{type}", getSensorData).Methods("GET")
	http.ListenAndServe(port, handlers.CORS(allowedHeaders, allowedOrigins, allowedMethods)(router))

}
