package agent

import (
	"encoding/json"
	"log"
	"net/http"
)

// JSONRequest ...
type JSONRequest struct {
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
	ID     interface{}   `json:"id"`
}

// JSONResponse ...
type JSONResponse struct {
	Result map[string]interface{} `json:"result"`
	Error  map[string]interface{} `json:"error"`
	ID     interface{}            `json:"id"`
}

// APIFunc ...
type APIFunc func(*JSONRequest) (*JSONResponse, error)

// ValidatorFunc ...
type ValidatorFunc func(*http.Request) int

// Handler ...
type Handler struct {
	handlerMap map[string]APIFunc
	validator  ValidatorFunc
}

// NewHandler ...
func NewHandler() *Handler {
	return &Handler{handlerMap: make(map[string]APIFunc)}
}

func (h Handler) ServeHTTP(s http.ResponseWriter, r *http.Request) {
	if h.validator != nil {
		ret := h.validator(r)
		if ret != 200 {
			s.WriteHeader(ret)
			return
		}
	}
	req := new(JSONRequest)
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		s.WriteHeader(400)
		log.Printf("Bad request: %s", err)
		return
	}
	api, ok := h.handlerMap[req.Method]
	if ok == false {
		api, ok = h.handlerMap["*"]
		if ok == false {
			s.WriteHeader(400)
			log.Printf("Method %s not registered", req.Method)
			return
		}
	}
	resp, aerr := api(req)
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
	h.handlerMap[method] = api
}

// RegisterWildCardAPI ...
func (h *Handler) RegisterWildCardAPI(api APIFunc) {
	h.handlerMap["*"] = api
}

// RegisterValidator ...
func (h *Handler) RegisterValidator(v ValidatorFunc) {
	h.validator = v
}
