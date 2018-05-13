package mcservice

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/vharish836/middleman/boxer"
)

// Config ...
type Config struct {
	UserName     string
	PassWord     string
	ChainName    string
	RPCPort      int
	RPCUser      string
	RPCPassword  string
	NativeEntity string
}

//MCService ...
type MCService struct {
	cfg   *Config
	boxer *boxer.Boxer
}

// NewService ...
func NewService(cfg *Config, b *boxer.Boxer) (*MCService, error) {
	s := MCService{
		cfg:   cfg,
		boxer: b,
	}
	err := s.initialize()
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// BuildAPIContext ...
func (s MCService) BuildAPIContext(r *http.Request, w http.ResponseWriter) (context.Context, error) {
	ret := s.checkAuth(r)
	if ret != 200 {
		w.WriteHeader(ret)
		return nil, errInternal
	}
	req := JSONRequest{}
	dec := json.NewDecoder(r.Body)
	dec.UseNumber()
	err := dec.Decode(&req)
	if err != nil {
		w.WriteHeader(400)
		log.Printf("Bad request: %s", err)
		return nil, err
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, reqKey, &req)
	return ctx, nil
}

// APIHandler ...
func (s MCService) APIHandler(ctx context.Context) context.Context {
	req := ctx.Value(reqKey).(*JSONRequest)
	api, ok := apiMap.Load(req.Method)
	if ok == false {
		api, ok = apiMap.Load("*")
	}
	method := api.(apiFunc)
	resp, err := method(req)
	ctx = context.WithValue(ctx, errKey, err)
	if err == nil {
		ctx = context.WithValue(ctx, rspKey, resp)
	}
	return ctx
}

// WriteResponse ...
func (s MCService) WriteResponse(ctx context.Context, w http.ResponseWriter) {
	req := ctx.Value(reqKey).(*JSONRequest)
	err := ctx.Value(errKey)
	switch err.(type) {
	case nil:
		resp := ctx.Value(rspKey).(*JSONResponse)
		resp.ID = req.ID
		writeResponse(resp, w)
	case error:
		e := err.(error)
		eresp := JSONResponse{Error: map[string]string{
			"error": e.Error(),
		}, ID: req.ID}
		writeResponse(&eresp, w)
		log.Printf("failed to handle API: %s", e)
	}
}
