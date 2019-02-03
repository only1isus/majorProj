package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

var userTestData = []struct {
	username       string
	password       string
	endpoint       string
	endpointName   string
	succesAuthResp int
	succesGetResp  int
}{
	{"test123@gmai.com", "password", "api/sensor/?sensortype=temperature&timespan=1", "sensor", http.StatusUnauthorized, http.StatusOK},
	{"test123@gmail.com", "password", "api/sensor/?sensortype=temperature&timespan=1", "sensor", http.StatusUnauthorized, http.StatusOK},
	{"test123@gmail.com", "qwerty", "api/sensor/?sensortypetemperature&timespan=1", "sensor", http.StatusOK, http.StatusNotFound},
	{"test123@gmail.com", "qwerty", "api/sensor/?sensortype=temperature&timespan=1", "sensor", http.StatusOK, http.StatusOK},
	{"test123@gmail.com", "qwerty", "api/logs/?timespan=1", "logs", http.StatusOK, http.StatusOK},
	{"test123@gmail.com", "qwerty", "api/logs/?timespan=e", "logs", http.StatusOK, http.StatusBadRequest},
	{"", "qwerty", "api/sensor/?sensortype=%s&timespan=1", "sensor", http.StatusUnauthorized, http.StatusBadRequest},
	{"", "", "api/sensor/?sensortype=humidity&timespan=1", "sensor", http.StatusUnauthorized, http.StatusBadRequest},
	{"", "", "userinfo", "userinfo", http.StatusUnauthorized, http.StatusOK},
	{"rom@gmail.com", "", "userinfo", "userinfo", http.StatusUnauthorized, http.StatusOK},
	{"", "password", "userinfo", "userinfo", http.StatusUnauthorized, http.StatusOK},
	{"test123@gmail.com", "qwerty", "userinfo", "userinfo", http.StatusOK, http.StatusOK},
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

func TestEndpoints(t *testing.T) {
	for _, user := range userTestData {
		t.Run(user.username, func(t *testing.T) {
			jwtToken, code, _ := authenticate(user.username, user.password)

			if code != user.succesAuthResp {
				t.Errorf("got %v instead of %v. endpoint %s.", code, user.succesAuthResp, user.endpoint)
			}
			if jwtToken != "" {
				req := httptest.NewRequest("GET", fmt.Sprintf("http://192.168.0.18:8080/%s", user.endpoint), nil)
				req.Header.Add("Token", jwtToken)
				w := httptest.NewRecorder()
				if user.endpointName == "logs" {
					getLogs(w, req)
				}
				if user.endpointName == "userinfo" {
					userinfo(w, req)
				}
				if user.endpointName == "sensor" {
					getSensorData(w, req)
				}
				resp := w.Result()
				if resp.StatusCode != user.succesGetResp {
					t.Errorf("got status code %v instead of %v while requesting %v", resp.StatusCode, user.succesGetResp, user.endpoint)
				}
			}

		})
	}
}
func TestRegister(t *testing.T) {
	data := map[string]string{
		"name":     "jon doe",
		"email":    "test@gmail.com",
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
		t.Errorf("got %v instead", w.Code)
	}
}
