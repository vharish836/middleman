package mcservice

import (
	"encoding/hex"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"sync"
	"github.com/vharish836/middleman/config"
	"github.com/vharish836/middleman/handler"
)

// Service ...
type Service struct {
	cfg        *config.Config
	h          *handler.Handler
	entityKeys sync.Map
	nativeKey  []byte
}

// NewService ...
func NewService(cfg *config.Config) *Service {
	return &Service{cfg: cfg}
}

// LoadEntityMap ...
func (s *Service) LoadEntityMap() error {
	if s.cfg.Keys == nil {
		return errors.New("no keys configured")
	}
	for i := range s.cfg.Keys {
		key,err := hex.DecodeString(s.cfg.Keys[i].Value)
		if err != nil {
			return err
		}
		s.entityKeys.Store(s.cfg.Keys[i].ID, string(key))
	}
	nativekey, ok := s.entityKeys.Load(s.cfg.NativeEntity)
	if ok != true {
		return errors.New("native entity key not configured")
	}
	s.nativeKey = []byte(nativekey.(string))
	return nil
}

//go:generate go run gen.go

// Initialize ...
func (s *Service) Initialize() (*handler.Handler, error) {
	s.h = handler.NewHandler()
	s.h.RegisterValidator(s.CheckAuth)
	s.RegisterAllAPI()
	err := s.LoadEntityMap()
	if err != nil {
		return nil, err
	}
	return s.h, nil
}

// CheckAuth ...
func (s *Service) CheckAuth(r *http.Request) int {
	username, password, ok := r.BasicAuth()
	if ok != true {
		return 401
	}
	if username != s.cfg.UserName || password != s.cfg.PassWord {
		return 403
	}
	return 200
}

// PlatformAPI ...
func (s *Service) PlatformAPI(req *handler.JSONRequest) (*handler.JSONResponse, error) {
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
	resp := handler.JSONResponse{}
	err = json.NewDecoder(rsp.Body).Decode(&resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// PassThru ...
func (s *Service) PassThru(req *handler.JSONRequest) (*handler.JSONResponse, error) {
	return s.PlatformAPI(req)
}
