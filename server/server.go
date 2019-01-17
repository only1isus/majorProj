package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/only1isus/majorProj/controller"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/only1isus/majorProj/consts"
	db "github.com/only1isus/majorProj/server/database"
	"github.com/only1isus/majorProj/types"
	"google.golang.org/grpc"
)

const (
	host     string = "localhost"
	port     string = ":8080"
	grpcPort string = ":8001"
)

func postUser(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Println(err)
	}

	var u types.User
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("could not parse the data sent"))
	}

	if err := json.Unmarshal(body, &u); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("something went wrong while getting the information"))

	}
	u.CreatedAt = time.Now().Unix()
	err = db.AddEntry(consts.User, []byte(u.Email), u)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)

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
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("please check the timespan parameter passed"))
		return
	}
	switch strings.ToLower(sensorType) {

	case "humidity":
		data, err := db.GetSensorData(consts.Humidity, int64(s))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("something went wrong trying to get the data requested"))
			return
		}
		sendResponse(w, data)
	case "temperature":
		data, err := db.GetSensorData(consts.Temperature, int64(s))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("something went wrong trying to get the data requested"))
			return
		}
		sendResponse(w, data)
	case "all":
		data, err := db.GetSensorData(consts.All, int64(s))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("something went wrong trying to get the data requested"))
			return
		}
		sendResponse(w, data)
	default:
		w.WriteHeader(http.StatusOK)
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
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("please check the timespan parameter passed"))
		return
	}
	logs, err := db.GetLogs(int64(s))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
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

type svr struct{}

func (s *svr) CommitSensorData(ctx context.Context, data *controller.SensorData) (*controller.SuccessResponse, error) {
	d := types.SensorEntry{}
	err := json.Unmarshal(data.Data, &d)
	if err != nil {
		fmt.Println(err)
		return &controller.SuccessResponse{Success: false}, err
	}
	err = db.AddEntry(consts.Sensor, []byte(time.Now().Format(time.RFC3339)), d)
	if err != nil {
		fmt.Println(err)
		return &controller.SuccessResponse{Success: false}, err
	}
	return &controller.SuccessResponse{Success: true}, nil
}

func (s *svr) CommitLog(ctx context.Context, data *controller.LogData) (*controller.SuccessResponse, error) {
	l := types.LogEntry{}
	if err := json.Unmarshal(data.Data, &l); err != nil {
		fmt.Println(err)
		return &controller.SuccessResponse{Success: false}, err
	}
	err := db.AddEntry(consts.Log, []byte(time.Now().Format(time.RFC3339)), l)
	if err != nil {
		return &controller.SuccessResponse{Success: false}, err
	}
	return &controller.SuccessResponse{Success: true}, nil
}

func main() {
	kill := make(chan bool)
	allowedHeaders := handlers.AllowedHeaders([]string{"application/json", "application/x-www-form-urlencoded", "multipart/form-data"})
	allowedOrigins := handlers.AllowedOrigins([]string{"*"})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	router := mux.NewRouter()
	log.Printf("server running pn port %s...", port)
	router.HandleFunc("/newuser", postUser).Methods("POST")
	router.HandleFunc("/api/sensor/", getSensorData).Methods("GET")
	router.HandleFunc("/api/logs/", getLogs).Methods("GET")
	router.HandleFunc("/api/settings", changeSettings).Methods("POST")
	go func() {
		err := http.ListenAndServe(port, handlers.CORS(allowedHeaders, allowedOrigins, allowedMethods)(router))
		if err != nil {
			log.Fatalf("cannot create a connection on port %s", port)
		}
	}()

	// make setting up a gRPC connection
	grpcsrv := grpc.NewServer()
	controller.RegisterCommitServer(grpcsrv, &svr{})

	conn, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Printf("cannot create a connection on port %s", grpcPort)
		os.Exit(1)
	}

	// run the server as a goroutine as to avoid blocking
	go func() {
		fmt.Printf("gRPC server running on port %v", grpcPort)
		if err := grpcsrv.Serve(conn); err != nil {
			log.Fatalf("failed to create gRPC serve: %v", err)
			os.Exit(1)
		}
	}()

	<-kill
}
