package mcservice

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/vharish836/middleman/config"
	"github.com/vharish836/middleman/handler"
)

// Service ...
type Service struct {
	cfg          *config.Config
	h            *handler.Handler
	entityKeys   *cache.Cache
	nativeEntity string
}

func (s *Service) loadEntityMap() error {
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

func (s *Service) initialize() error {
	s.h = handler.NewHandler()
	s.h.RegisterValidator(s.CheckAuth)
	s.RegisterAllAPI()
	err := s.loadEntityMap()
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) platformAPI(req *handler.JSONRequest) (*handler.JSONResponse, error) {
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

// PassThru ...
func (s *Service) PassThru(req *handler.JSONRequest) (*handler.JSONResponse, error) {
	return s.platformAPI(req)
}

// GetHandler ...
func (s *Service) GetHandler() *handler.Handler {
	return s.h
}

// NewService ...
func NewService(cfg *config.Config) (*Service, error) {
	d, err := time.ParseDuration(cfg.Cache.TTL)
	if err != nil {
		return nil, err
	}
	c, err := time.ParseDuration(cfg.Cache.CleanupInterval)
	if err != nil {
		return nil, err
	}
	ch := cache.New(d, c)
	ch.OnEvicted(func(k string, v interface{}) {
		log.Printf("key: \"%s\" got evicted", k)
	})
	s := &Service{
		cfg:        cfg,
		entityKeys: ch,
	}
	err = s.initialize()
	if err != nil {
		return nil, err
	}
	return s, nil
}
