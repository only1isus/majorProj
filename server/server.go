package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/only1isus/majorProj/consts"
	"github.com/only1isus/majorProj/database"
	"github.com/only1isus/majorProj/types"
)

const (
	host string = "localhost"
	port string = ":8080"
)

func postUser(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Println(err)
	}

	var u types.User
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("could not parse the data sent"))
	}

	if err := json.Unmarshal(body, &u); err != nil {
		w.WriteHeader(500)
		w.Write([]byte("something went wrong while getting the information"))

	}
	u.CreatedAt = time.Now().Unix()
	err = db.AddEntry(consts.User, []byte(u.Email), u)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("something went wrong while getting the information"))
		fmt.Println(err)
	} else {
		w.Header()
		w.WriteHeader(200)
	}
}

func sendResponse(w http.ResponseWriter, data interface{}) {
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

// returns the sensor data by type
func getSensorData(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	span := query.Get("timespan")
	sensorType := query.Get("sensortype")
	s, err := strconv.Atoi(span)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("please check the timespan parameter passed"))
		return
	}
	switch sensorType {
	case "temperature":
		data, err := db.GetSensorData(consts.Temperature, int64(s))
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("something went wrong trying to get the data requested"))
			return
		}
		sendResponse(w, data)
	case "all":
		data, err := db.GetSensorData(consts.All, int64(s))
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("something went wrong trying to get the data requested"))
			return
		}
		sendResponse(w, data)
	default:
		w.WriteHeader(404)
		w.Write([]byte("404 page not found"))
		return
	}
}

// get all the logs
func getLogs(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	span := query.Get("timespan")
	s, err := strconv.Atoi(span)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("please check the timespan parameter passed"))
		return
	}
	logs, err := db.GetLogs(int64(s))
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("something went wrong trying to get the data requested"))
	}
	sendResponse(w, logs)
}

// changeSettings edits the config file of the system
func changeSettings(w http.ResponseWriter, r *http.Request) {
	fmt.Println("trying to change settings")
	if err := r.ParseForm(); err != nil {
		fmt.Println(err)
	}
}

func main() {
	allowedHeaders := handlers.AllowedHeaders([]string{"application/json", "application/x-www-form-urlencoded", "multipart/form-data"})
	allowedOrigins := handlers.AllowedOrigins([]string{"*"})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	router := mux.NewRouter()
	log.Printf("server running pn port %s...", port)
	router.HandleFunc("/newuser", postUser).Methods("POST")
	router.HandleFunc("/api/sensor/", getSensorData).Methods("GET")
	router.HandleFunc("/api/logs/", getLogs).Methods("GET")
	router.HandleFunc("/api/settings", changeSettings).Methods("POST")
	http.ListenAndServe(port, handlers.CORS(allowedHeaders, allowedOrigins, allowedMethods)(router))
}
