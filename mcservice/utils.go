package mcservice

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"sync"

	cache "github.com/patrickmn/go-cache"
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

func (s *MCService) loadEntityMap() error {
	if s.cfg.Crypto.Keys == nil {
		return errors.New("no keys configured")
	}
	for i := range s.cfg.Crypto.Keys {
		key, err := hex.DecodeString(s.cfg.Crypto.Keys[i].Value)
		if err != nil {
			return err
		}
		if s.cfg.Crypto.Keys[i].Native {
			err = s.entityKeys.Add(s.cfg.Crypto.Keys[i].ID, key, cache.NoExpiration)
			if err != nil {
				return err
			}
			if s.nativeEntity == "" {
				s.nativeEntity = s.cfg.Crypto.Keys[i].ID
			} else {
				return errors.New("only one key can be marked native")
			}
		} else {
			s.entityKeys.SetDefault(s.cfg.Crypto.Keys[i].ID, key)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

//go:generate go run gen.go

func (s *MCService) initialize() error {
	s.registerAllAPI()
	err := s.loadEntityMap()
	if err != nil {
		return err
	}
	return nil
}

func (s *MCService) platformAPI(req *JSONRequest) (*JSONResponse, error) {
	rbuf, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	r, nerr := http.NewRequest("POST",
		"http://localhost:"+strconv.Itoa(s.cfg.MultiChain.RPCPort)+"/", bytes.NewBuffer(rbuf))
	if nerr != nil {
		return nil, nerr
	}
	r.SetBasicAuth(s.cfg.MultiChain.RPCUser, s.cfg.MultiChain.RPCPassword)
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
