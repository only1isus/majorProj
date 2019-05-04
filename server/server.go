package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/only1isus/majorProj/config"

	"github.com/joho/godotenv"
	"github.com/segmentio/ksuid"

	"golang.org/x/crypto/bcrypt"

	"github.com/only1isus/majorProj/controller"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/only1isus/majorProj/consts"
	db "github.com/only1isus/majorProj/server/database"

	"github.com/only1isus/majorProj/rpc"
	"github.com/only1isus/majorProj/types"
	"google.golang.org/grpc"
)

const (
	port string = ":8080"
)

// hard coded for testing reasons. ENV will be used eventually
func getSecret() ([]byte, error) {
	if err := godotenv.Load(".env"); err != nil {
		return nil, err
	}
	return []byte(os.Getenv("SIGKEY")), nil
}

func register(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Errorf("something went wrong parsing body"))
		return
	}

	var u types.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Errorf("something went wrong parsing the json data"))
		return
	}

	pword, err := hashPassword(u.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	// formatting the data to be saved in the database.
	u.Phone = strings.TrimSpace(u.Phone)
	u.Email = strings.ToLower(u.Email)
	u.Password = string(pword)
	u.CreatedAt = time.Now().Unix()
	key := ksuid.New()
	u.Key = strings.ToUpper(key.String())

	err = db.AddUserEntry(u)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Errorf("something went wrong %v ", err.Error()))
	}
	if err := db.CreateBucket(u.Key); err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Errorf("something went wrong creating bucket %v ", err.Error()))
	}
	return
}

func sendResponse(w http.ResponseWriter, data interface{}) {
	d, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(d)
	}
	return
}

func respondWithError(w http.ResponseWriter, statusCode int, err error) {
	w.WriteHeader(statusCode)
	w.Write([]byte(err.Error()))
	return
}

func getClaims(w http.ResponseWriter, r *http.Request) jwt.MapClaims {
	token, err := jwt.Parse(r.Header["Token"][0], func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("There was an error")
		}

		sigKey, err := getSecret()
		if err != nil {
			return "", fmt.Errorf("Something Went Wrong: %s", err.Error())
		}
		return sigKey, nil
	})
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, fmt.Errorf("jwt invalid"))
	}
	claims := token.Claims.(jwt.MapClaims)
	return claims
}

// returns the sensor data by type
func getSensorData(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	s := query.Get("starttime")
	e := query.Get("endtime")
	sensorType := query.Get("sensortype")
	if s == "" || e == "" || sensorType == "" {
		respondWithError(w, http.StatusBadRequest, fmt.Errorf("empty parameters being passed"))
		return
	}
	start, err := strconv.Atoi(s)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Errorf("check the value of parameters being passed"))
		return
	}
	end, err := strconv.Atoi(e)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Errorf("check the value of parameters being passed"))
		return
	}
	claims := getClaims(w, r)
	var st consts.BucketFilter
	key := claims["key"].(string)
	switch strings.ToLower(sensorType) {
	case "humidity":
		st = consts.Humidity
	case "temperature":
		st = consts.Temperature
	case "waterlevel":
		st = consts.WaterLevel
	case "ph":
		st = consts.PH
	case "all":
		st = consts.All
	default:
		respondWithError(w, http.StatusNotFound, fmt.Errorf("page not found"))
		return
	}

	data, err := db.GetSensorData([]byte(key), st, int64(start), int64(end))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}
	sendResponse(w, data)
	return
}

// get all the logs
func getLogs(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	s := query.Get("starttime")
	e := query.Get("endtime")
	if s == "" || e == "" {
		respondWithError(w, http.StatusBadRequest, fmt.Errorf("empty parameters being passed"))
		return
	}
	start, err := strconv.Atoi(s)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Errorf("check the value of parameters being passed"))
		return
	}
	end, err := strconv.Atoi(e)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Errorf("check the value of parameters being passed"))
		return
	}

	claims := getClaims(w, r)
	key := claims["key"].(string)
	logs, err := db.GetLogs([]byte(key), int64(start), int64(end))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Errorf("Something went wrong getting the data requested"))
		return
	}
	sendResponse(w, logs)
	return
}

// changeSettings edits the config file of the system
func changeSettings(w http.ResponseWriter, r *http.Request) {
	fmt.Println("trying to change settings")
	if err := r.ParseForm(); err != nil {
		fmt.Println(err)
	}
}

func userinfo(w http.ResponseWriter, r *http.Request) {
	claims := getClaims(w, r)
	email := claims["client"].(string)
	u, err := db.GetUserData(email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, fmt.Errorf("user not found"))
		return
	}
	userInfo := types.User{
		Key:       (*u).Key,
		Name:      (*u).Name,
		Email:     (*u).Email,
		Phone:     (*u).Phone,
		CreatedAt: (*u).CreatedAt,
	}
	sendResponse(w, &userInfo)
	return
}

func addFarmDetails(w http.ResponseWriter, r *http.Request) {
	claims := getClaims(w, r)
	key := claims["key"].(string)
	fd := types.FarmDetails{}

	var farmDetails map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&farmDetails); err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Errorf("something went wrong decoding the data %v", err))
		return
	}

	harvDate, _ := farmDetails["harvestOn"].(float64)
	plantDate, _ := farmDetails["plantedOn"].(float64)

	fd.Configured = true
	fd.CropType = farmDetails["cropType"].(string)
	fd.HarvestOn = int64(harvDate)
	fd.PlantedOn = int64(plantDate)
	fd.NPK = farmDetails["npk"].(string)

	if err := db.AddFarmEntry([]byte(key), []byte(key), fd); err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}
	return
}

func getFarmDetails(w http.ResponseWriter, r *http.Request) {
	claims := getClaims(w, r)
	key := claims["key"].(string)
	fd, err := db.GetFarmDetails([]byte(key))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}
	sendResponse(w, *fd)
	return
}

func generateSummary(w http.ResponseWriter, r *http.Request) {
	s := new(types.Summary)
	claims := getClaims(w, r)
	key := claims["key"].(string)

	fd, err := db.GetFarmDetails([]byte(key))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	sensorData, err := db.GetSensorData([]byte(key), consts.All, fd.PlantedOn, fd.HarvestOn)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	s.ID = fmt.Sprintf("%d-%d", fd.PlantedOn, fd.HarvestOn)
	s.FarmDetails = *fd
	planted := time.Unix(fd.PlantedOn, fd.PlantedOn*1000)
	harvested := time.Unix(fd.HarvestOn, fd.HarvestOn*1000)

	var weekStart time.Time
	var weekEnd time.Time
	totalDays := int(math.Floor(harvested.Sub(planted).Hours() / 24))
	remDays := totalDays % 7

	for week := 1; week < ((totalDays-remDays)/7)+1; week++ {
		weekEntry := new(types.Week)

		weekStart = planted.AddDate(0, 0, 7*(week-1))
		weekEnd = planted.AddDate(0, 0, 7*week)
		weekEntry.WeekOf.Start = weekStart.Unix()
		weekEntry.WeekOf.End = weekEnd.Unix()

		for _, entry := range *sensorData {
			filterSensorEntry(entry, weekEntry, weekStart, weekEnd)
		}
		s.Data = append(s.Data, *weekEntry)
	}

	weekEntry := new(types.Week)
	weekEntry.WeekOf.Start = planted.AddDate(0, 0, totalDays-remDays).Unix()
	weekEntry.WeekOf.End = harvested.Unix()
	for _, entry := range *sensorData {
		filterSensorEntry(entry, weekEntry, planted.AddDate(0, 0, totalDays-remDays), harvested)
	}
	s.Data = append(s.Data, *weekEntry)

	if err := db.AddSummary(*s); err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return
}

func getsummaries(w http.ResponseWriter, r *http.Request) {
	summaries, err := db.GetSummaries()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}
	sendResponse(w, *summaries)
	return
}

func getToken(w http.ResponseWriter, r *http.Request) {
	email, password, _ := r.BasicAuth()
	if email == "" || password == "" {
		respondWithError(w, http.StatusUnauthorized, fmt.Errorf("please add username and password"))
		return
	}
	user, err := db.GetUserData(email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	if isSame := comparePasswords([]byte((*user).Password), []byte(password)); !isSame {
		respondWithError(w, http.StatusUnauthorized, fmt.Errorf("password or username not correct"))
		return
	}
	validToken, err := generateToken(user)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, fmt.Errorf("not authorized"))
		return
	}
	sendResponse(w, validToken)
	return
}

func generateToken(u *types.User) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	claims["client"] = (*u).Email
	claims["key"] = (*u).Key
	claims["exp"] = time.Now().Add(time.Hour * 48).Unix()
	sigKey, err := getSecret()
	if err != nil {
		return "", fmt.Errorf("Something Went Wrong: %s", err.Error())
	}
	tokenString, err := token.SignedString(sigKey)
	if err != nil {
		return "", fmt.Errorf("Something Went Wrong: %s", err.Error())
	}
	return tokenString, nil
}

func isProtected(endpoint func(http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header["Token"] != nil {
			if r.Header["Token"][0] == "" {
				respondWithError(w, http.StatusUnauthorized, fmt.Errorf("No token provided"))
				return
			}
			token, err := jwt.Parse(r.Header["Token"][0], func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("failed to authenticate")
				}
				sigKey, err := getSecret()
				if err != nil {
					return "", fmt.Errorf("Something Went Wrong: %s", err.Error())
				}
				return sigKey, nil
			})
			if err != nil {
				respondWithError(w, http.StatusUnauthorized, err)
				return
			}
			if token.Valid {
				claims := token.Claims.(jwt.MapClaims)
				email := claims["client"].(string)
				user, err := db.GetUserData(email)
				if err != nil {
					respondWithError(w, http.StatusUnauthorized, fmt.Errorf("trouble verifying user credentials"))
					return
				}
				if strings.ToLower(user.Email) != email {
					respondWithError(w, http.StatusUnauthorized, fmt.Errorf("invalid token"))
					return
				}
				endpoint(w, r)
				return
			}
		} else {
			respondWithError(w, http.StatusUnauthorized, fmt.Errorf("No token header"))
			return
		}
	})
}

func filterSensorEntry(e types.SensorEntry, w *types.Week, weekStart time.Time, weekEnd time.Time) {
	if e.Time >= weekStart.Unix() && e.Time <= weekEnd.Unix() {
		switch e.SensorType {
		case consts.Temperature:
			w.Data.Temperature.Values = append(w.Data.Temperature.Values, e.Value)
		case consts.WaterLevel:
			w.Data.WaterLevel.Values = append(w.Data.WaterLevel.Values, e.Value)
		}
	}
}

func hashPassword(password string) ([]byte, error) {
	p, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func comparePasswords(hashedPassword, password []byte) bool {
	if err := bcrypt.CompareHashAndPassword(hashedPassword, password); err != nil {
		return false
	}
	return true
}

func server() *http.Server {
	allowedHeaders := handlers.AllowedHeaders([]string{"application/json", "application/x-www-form-urlencoded", "Origin", "Access-Control-Allow-Origin", "X-Requested-With", "Content-Type", "Accept", "multipart/form-data", "Token", "Authorization"})
	allowedOrigins := handlers.AllowedOrigins([]string{"*"})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})
	router := mux.NewRouter()

	log.Printf("server running pn port %s...", port)
	router.HandleFunc("/register", register).Methods("POST")
	router.HandleFunc("/token", getToken).Methods("GET")
	router.Handle("/api/sensor/", isProtected(getSensorData)).Methods("GET")
	router.Handle("/userinfo", isProtected(userinfo)).Methods("GET")
	router.Handle("/api/logs/", isProtected(getLogs)).Methods("GET")
	router.Handle("/api/settings", isProtected(changeSettings)).Methods("POST")
	router.Handle("/api/farmdetails", isProtected(addFarmDetails)).Methods("POST")
	router.Handle("/api/farmdetails", isProtected(getFarmDetails)).Methods("GET")
	router.Handle("/api/generatesummary", isProtected(generateSummary)).Methods("GET")
	router.Handle("/api/getsummaries", isProtected(getsummaries)).Methods("GET")
	return &http.Server{
		Addr:    ":8080",
		Handler: handlers.CORS(allowedHeaders, allowedOrigins, allowedMethods)(router),
	}
}

func main() {
	conf, err := config.ReadConfigFile()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	c := types.Database{}
	if err := yaml.Unmarshal(conf, &c); err != nil {
		log.Printf("cannot get the notification setting: %v\n", err)
		os.Exit(1)
	}

	kill := make(chan bool)

	go func() {
		server := server()
		err := http.ListenAndServe(server.Addr, server.Handler)
		if err != nil {
			log.Fatalf("cannot create a connection on port %s\n", port)
		}
	}()

	grpcsrv := grpc.NewServer()
	controller.RegisterCommitServer(grpcsrv, &rpc.CommitSVR{})

	go rpc.NewServer(grpcsrv, fmt.Sprintf("%s:%s", c.Connection.Host, c.Connection.Port))

	<-kill
}
