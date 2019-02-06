package main

import (
	"context"
	"crypto/tls"
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

	"github.com/joho/godotenv"
	"github.com/segmentio/ksuid"

	"golang.org/x/crypto/bcrypt"

	"github.com/only1isus/majorProj/controller"

	jwt "github.com/dgrijalva/jwt-go"
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
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Errorf("something went wrong parsing the json data"))
		return
	}

	if err := json.Unmarshal(body, &u); err != nil {
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

	err = db.AddEntry(nil, consts.User, []byte(u.Email), u)
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
		fmt.Printf("got an error while trying encode response")
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
	span := query.Get("timespan")
	sensorType := query.Get("sensortype")
	s, err := strconv.Atoi(span)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Errorf("check the parameters being passed"))
	}

	claims := getClaims(w, r)
	key := claims["key"].(string)
	switch strings.ToLower(sensorType) {

	case "humidity":
		data, err := db.GetSensorData([]byte(key), consts.Humidity, int64(s))
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err)
			return
		}
		sendResponse(w, data)
		return
	case "temperature":
		data, err := db.GetSensorData([]byte(key), consts.Temperature, int64(s))
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err)
			return
		}
		sendResponse(w, data)
		return
	case "all":
		data, err := db.GetSensorData([]byte(key), consts.All, int64(s))
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err)
			return
		}
		sendResponse(w, data)
		return
	default:
		respondWithError(w, http.StatusNotFound, fmt.Errorf("page not found"))
		return
	}
}

// get all the logs
func getLogs(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	span := query.Get("timespan")
	s, err := strconv.Atoi(span)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Errorf("make sure the timespan is an integer"))
		return
	}

	claims := getClaims(w, r)
	key := claims["key"].(string)
	logs, err := db.GetLogs([]byte(key), int64(s))
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
				respondWithError(w, http.StatusUnauthorized, fmt.Errorf("not authorized"))
				return
			}
			token, err := jwt.Parse(r.Header["Token"][0], func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					respondWithError(w, http.StatusUnauthorized, fmt.Errorf("failed to authenticate"))
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
				endpoint(w, r)
			}
		} else {
			respondWithError(w, http.StatusUnauthorized, fmt.Errorf("not authorized"))
			return
		}
	})
}

func hashPassword(password string) ([]byte, error) {
	p, err := bcrypt.GenerateFromPassword([]byte(password), 14)
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

type svr struct{}

func (s *svr) CommitSensorData(ctx context.Context, data *controller.SensorData) (*controller.SuccessResponse, error) {
	d := types.SensorEntry{}
	err := json.Unmarshal(data.Data, &d)
	if err != nil {
		return &controller.SuccessResponse{Success: false}, err
	}
	err = db.AddEntry(data.Key, consts.Sensor, []byte(time.Now().Format(time.RFC3339)), d)
	if err != nil {
		return &controller.SuccessResponse{Success: false}, err
	}
	return &controller.SuccessResponse{Success: true}, nil
}

func (s *svr) CommitLog(ctx context.Context, data *controller.LogData) (*controller.SuccessResponse, error) {
	l := types.LogEntry{}
	if err := json.Unmarshal(data.Data, &l); err != nil {
		return &controller.SuccessResponse{Success: false}, err
	}
	err := db.AddEntry(data.Key, consts.Log, []byte(time.Now().Format(time.RFC3339)), l)
	if err != nil {
		return &controller.SuccessResponse{Success: false}, err
	}
	return &controller.SuccessResponse{Success: true}, nil
}

func certReqFunc(certfile, keyfile string) func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	c, err := getCert(certfile, keyfile)
	return func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		if err != nil {
			return nil, err
		}
		return &c, nil
	}
}

func clientCertReqFunc(certfile, keyfile string) func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
	c, err := getCert(certfile, keyfile)

	return func(certReq *tls.CertificateRequestInfo) (*tls.Certificate, error) {
		if err != nil || certfile == "" {
			return nil, err
		}
		return &c, nil
	}
}

func getCert(certfile, keyfile string) (c tls.Certificate, err error) {
	if certfile != "" && keyfile != "" {
		c, err = tls.LoadX509KeyPair(certfile, keyfile)
		if err != nil {
			fmt.Printf("Error loading key pair: %v\n", err)
		}
	} else {
		err = fmt.Errorf("I have no certificate")
	}
	return
}

func main() {
	kill := make(chan bool)
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
	router.HandleFunc("/api/settings", changeSettings).Methods("POST")

	go func() {
		// cp, _ := x509.SystemCertPool()
		// data, _ := ioutil.ReadFile("minica.pem")
		// cp.AppendCertsFromPEM(data)

		// tls := &tls.Config{
		// 	ClientCAs:      cp,
		// 	GetCertificate: certReqFunc("./192.168.0.18/cert.pem", "./192.168.0.18/key.pem"),
		// }

		server := &http.Server{
			Addr: ":8080",
			// TLSConfig: tls,
			Handler: handlers.CORS(allowedHeaders, allowedOrigins, allowedMethods)(router),
		}

		err := server.ListenAndServe()
		if err != nil {
			log.Fatalf("cannot create a connection on port %s", port)
		}
	}()

	// make setting up a gRPC connection
	grpcsrv := grpc.NewServer()
	controller.RegisterCommitServer(grpcsrv, &svr{})

	conn, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Printf("cannot create a connection on port %s\n", grpcPort)
		os.Exit(1)
	}

	// run the server as a goroutine as to avoid blocking
	go func() {
		fmt.Printf("gRPC server running on port %v\n", grpcPort)
		if err := grpcsrv.Serve(conn); err != nil {
			log.Fatalf("failed to create gRPC serve: %v\n", err)
			os.Exit(1)
		}
	}()

	<-kill
}
