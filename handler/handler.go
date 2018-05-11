package handler

import (
	"encoding/json"
	"log"
	"net/http"
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

// APIFunc ...
type APIFunc func(*JSONRequest) (*JSONResponse, error)

// ValidatorFunc ...
type ValidatorFunc func(*http.Request) int

var handlerMap sync.Map

// Handler ...
type Handler struct {	
	validator  ValidatorFunc
}

// NewHandler ...
func NewHandler() *Handler {
	return &Handler{}
}

func (h Handler) ServeHTTP(s http.ResponseWriter, r *http.Request) {
	if h.validator != nil {
		ret := h.validator(r)
		if ret != 200 {
			s.WriteHeader(ret)
			return
		}
	}
	req := JSONRequest{}
	dec := json.NewDecoder(r.Body)
	dec.UseNumber()
	err := dec.Decode(&req)
	if err != nil {
		s.WriteHeader(400)
		log.Printf("Bad request: %s", err)
		return
	}
	api, ok := handlerMap.Load(req.Method)
	if ok == false {
		api, ok = handlerMap.Load("*")
		if ok == false {
			s.WriteHeader(400)
			log.Printf("Method %s not registered", req.Method)
			return
		}
	}
	method := api.(APIFunc)
	resp, aerr := method(&req)
	if aerr != nil {
		s.WriteHeader(500)
		log.Printf("failed to handle API: %s", aerr)
		return
	}
	resp.ID = req.ID
	rbuf, eerr := json.Marshal(resp)
	if eerr != nil {
		s.WriteHeader(500)
		log.Printf("could not encode response: %s", eerr)
		return
	}
	s.Header().Set("Content-Type", "application/json")
	_, err = s.Write(rbuf)
	if err != nil {
		log.Printf("could not write response: %s", err)
		return
	}
}

// RegisterAPI ...
func (h *Handler) RegisterAPI(method string, api APIFunc) {
	handlerMap.Store(method,api)
}

// RegisterWildCardAPI ...
func (h *Handler) RegisterWildCardAPI(api APIFunc) {
	handlerMap.Store("*",api)
}

// RegisterValidator ...
func (h *Handler) RegisterValidator(v ValidatorFunc) {
	h.validator = v
}
