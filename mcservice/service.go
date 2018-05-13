package mcservice

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
	"context"

	"github.com/patrickmn/go-cache"
	"github.com/vharish836/middleman/config"
)

//MCService ...
type MCService struct {
	cfg          *config.Config
	entityKeys   *cache.Cache
	nativeEntity string
}

// NewService ...
func NewService(cfg *config.Config) (*MCService, error) {
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
	s := MCService{
		cfg:        cfg,
		entityKeys: ch,
	}
	err = s.initialize()
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
