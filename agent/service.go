package agent

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// Service ...
type Service struct {
	cfg *Config
	h   *Handler
}

// NewService ...
func NewService(cfg *Config) *Service {
	return &Service{cfg: cfg}
}

// Initialize ...
func (s *Service) Initialize() (*Handler,error) {
	s.h = NewHandler()
	s.h.RegisterWildCardAPI(s.PassThru)
	s.h.RegisterValidator(s.CheckAuth)
	return s.h,nil
}

// CheckAuth ...
func (s *Service) CheckAuth(r *http.Request) int {
	username, password, ok := r.BasicAuth()
	if ok != true || username != s.cfg.UserName || password != s.cfg.PassWord {
		return 401
	}
	return 200
}

// PlatformAPI ...
func (s *Service) PlatformAPI(req *JSONRequest) (*JSONResponse, error) {
	rbuf, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	r, nerr := http.NewRequest("POST", "http://localhost:"+s.cfg.MultiChain.RPCPort+"/", bytes.NewBuffer(rbuf))
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
	resp := new(JSONResponse)
	err = json.NewDecoder(rsp.Body).Decode(resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// PassThru ...
func (s *Service) PassThru(req *JSONRequest) (*JSONResponse, error) {
	return s.PlatformAPI(req)
}
