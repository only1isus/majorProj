package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type auth struct {
	username string
	password string
	response int
}

type endpoint struct {
	reqType  string
	endpoint string
	name     string
	response int
}

var testData = []struct {
	userAuth            *auth
	endpointInformation endpoint
	omitToken           bool
	data                interface{}
}{

	// {
	// 	userAuth:            &auth{username: "test4@gmail.com", password: "password", response: http.StatusUnauthorized},
	// 	endpointInformation: endpoint{reqType: "get"},
	// },
	{
		userAuth:            &auth{username: "test1@gmail.com", password: "qwerty", response: http.StatusOK},
		endpointInformation: endpoint{endpoint: fmt.Sprintf("api/sensor/?sensortype=temperature&starttime=%d&endtime=%d", 1553299200, time.Now().Unix()), name: "sensor", response: http.StatusOK, reqType: "get"},
	},
	// {
	// 	userAuth:            &auth{username: "test4@gmail.com", password: "qwerty", response: http.StatusOK},
	// 	endpointInformation: endpoint{endpoint: "api/sensor/?sensortype=temperature&timepan=1", name: "sensor", response: http.StatusBadRequest, reqType: "get"},
	// 	omitToken:           false,
	// },
	// {
	// 	userAuth:            &auth{username: "test4@gmail.com", password: "qwerty", response: http.StatusOK},
	// 	endpointInformation: endpoint{endpoint: "api/farmdetails", name: "farmdetails", reqType: "post", response: http.StatusOK},
	// 	data: types.FarmDetails{
	// 		Configured:   true,
	// 		CropType:     "lettuce",
	// 		MaturityTime: 37,
	// 	},
	// },
	// {
	// 	userAuth:            &auth{username: "test4@gmail.com", password: "qwerty", response: http.StatusOK},
	// 	endpointInformation: endpoint{endpoint: "api/farmdetails", name: "farmdetails", response: http.StatusOK, reqType: "get"},
	// 	omitToken:           false,
	// },
}

func authenticate(username, password string) (string, int, error) {
	req := httptest.NewRequest("GET", "http://192.168.0.18:8080/token", nil)
	req.SetBasicAuth(username, password)
	w := httptest.NewRecorder()
	getToken(w, req)
	res := w.Result()
	if res.StatusCode != http.StatusOK {
		return "", res.StatusCode, fmt.Errorf("got %v intstead", res.StatusCode)
	}
	var token string
	err := json.NewDecoder(res.Body).Decode(&token)
	if err != nil {
		return "", res.StatusCode, err
	}
	return token, res.StatusCode, nil
}

func TestProtectedEndpoints(t *testing.T) {
	s := server()
	svr := httptest.NewServer(s.Handler)
	defer svr.Close()

	for _, td := range testData {
		t.Run(td.endpointInformation.name, func(t *testing.T) {

			if td.endpointInformation.reqType == "" {
				t.Fatal("no request type specified")
			}

			jwtToken, code, err := authenticate(td.userAuth.username, td.userAuth.password)
			if err != nil {
				t.Log("got an error trying to authenticate the user")
			}
			r := &http.Request{}
			if code != td.userAuth.response {
				t.Errorf("got %v instead of %v.", code, td.userAuth.response)
			}

			if td.endpointInformation.reqType == "get" {
				req := httptest.NewRequest("GET", fmt.Sprintf("%s%s", svr.URL, td.endpointInformation.endpoint), nil)
				req.Header.Add("Token", jwtToken)
				r = req
			}
			if td.endpointInformation.reqType == "post" {
				if td.data == nil {
					t.Fatalf("data needs to be added to run this test. endpoint %s", td.endpointInformation.endpoint)
				}
				out, err := json.Marshal(td.data)
				if err != nil {
					t.Error("couldn't marshal the json data")
				}
				req := httptest.NewRequest("POST", fmt.Sprintf("%s%s", svr.URL, td.endpointInformation.endpoint), bytes.NewReader(out))
				req.Header.Add("Token", jwtToken)
				req.Header.Set("Content-Type", "application/json")
				r = req
			}
			if jwtToken != "" {
				if td.omitToken {
					r.Header.Del("Token")
				}
				// h := &http.Response{}
				w := httptest.NewRecorder()
				if td.endpointInformation.reqType == "get" {
					h := http.HandlerFunc(
						isProtected(func(w http.ResponseWriter, r *http.Request) {
							if td.endpointInformation.name == "logs" {
								getLogs(w, r)
							}
							if td.endpointInformation.name == "userinfo" {
								userinfo(w, r)
							}
							if td.endpointInformation.name == "sensor" {
								getSensorData(w, r)
							}
							if td.endpointInformation.name == "farmdetails" {
								getFarmDetails(w, r)
							}
						}).(http.HandlerFunc),
					)
					h.ServeHTTP(w, r)
					resp := w.Result()
					if resp.StatusCode != td.endpointInformation.response {
						t.Errorf("got status code %v instead of %v while requesting %v", resp.StatusCode, td.endpointInformation.response, td.endpointInformation.name)
						b, err := ioutil.ReadAll(resp.Body)
						if err != nil {
							t.Log(err.Error())
						}
						t.Log(string(b))
					}
				}
				if td.endpointInformation.reqType == "post" {
					h := http.HandlerFunc(
						isProtected(func(w http.ResponseWriter, r *http.Request) {
							if td.endpointInformation.name == "farmdetails" {
								addFarmDetails(w, r)
							}
							if td.endpointInformation.name == "register" {
								register(w, r)
							}
						}).(http.HandlerFunc),
					)
					h.ServeHTTP(w, r)
					resp := w.Result()
					if resp.StatusCode != td.endpointInformation.response {
						t.Errorf("got status code %v instead of %v while requesting %v", resp.StatusCode, td.endpointInformation.response, td.endpointInformation.name)
						b, err := ioutil.ReadAll(resp.Body)
						if err != nil {
							t.Log(err.Error())
						}
						t.Log(string(b))
					}
				}
			}
		})
	}
}
func TestRegister(t *testing.T) {
	data := map[string]string{
		"name":     "jon doe",
		"email":    "test1@gmail.com",
		"phone":    "8785980103",
		"password": "qwerty",
	}
	out, err := json.Marshal(data)
	if err != nil {
		t.Error("couldn't marshal the json data")
	}

	req := httptest.NewRequest("POST", "http://192.168.0.18:8080/register", bytes.NewReader(out))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	register(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("got %v instead, %v", w.Code, w.Body.String())
	}
}
