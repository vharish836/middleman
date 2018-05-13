package mcservice

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
)

// JSONRequest ...
type JSONRequest struct {
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
	ID     interface{}   `json:"id"`
}

// JSONResponse ...
type JSONResponse struct {
	Result interface{} `json:"result"`
	Error  interface{} `json:"error"`
	ID     interface{} `json:"id"`
}

var apiMap sync.Map

type key int

const userKey key = 0
const reqKey key = 1
const rspKey key = 2
const errKey key = 3

// apiFunc ...
type apiFunc func(*JSONRequest) (*JSONResponse, error)

func writeResponse(resp *JSONResponse, s http.ResponseWriter) {
	rbuf, err := json.Marshal(resp)
	if err != nil {
		s.WriteHeader(500)
		log.Printf("could not encode response: %s", err)
		return
	}
	s.Header().Set("Content-Type", "application/json")
	_, err = s.Write(rbuf)
	if err != nil {
		log.Printf("could not write response: %s", err)
	}
	return
}

func (s *MCService) registerAPI(method string, api apiFunc) {
	apiMap.Store(method, api)
}

func (s *MCService) registerWildCardAPI(api apiFunc) {
	apiMap.Store("*", api)
}

//go:generate go run gen.go

func (s *MCService) initialize() error {
	s.registerAllAPI()
	return nil
}

func (s *MCService) platformAPI(req *JSONRequest) (*JSONResponse, error) {
	rbuf, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	r, nerr := http.NewRequest("POST",
		"http://localhost:"+strconv.Itoa(s.cfg.RPCPort)+"/", bytes.NewBuffer(rbuf))
	if nerr != nil {
		return nil, nerr
	}
	r.SetBasicAuth(s.cfg.RPCUser, s.cfg.RPCPassword)
	r.Header.Set("Content-Type", "application/json")
	rsp, derr := http.DefaultClient.Do(r)
	if derr != nil {
		return nil, derr
	}
	defer rsp.Body.Close()
	resp := JSONResponse{}
	err = json.NewDecoder(rsp.Body).Decode(&resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// checkAuth ...
func (s *MCService) checkAuth(r *http.Request) int {
	username, password, ok := r.BasicAuth()
	if ok != true {
		return 401
	}
	if username != s.cfg.UserName || password != s.cfg.PassWord {
		return 403
	}
	return 200
}

func (s *MCService) passThru(req *JSONRequest) (*JSONResponse, error) {
	return s.platformAPI(req)
}
