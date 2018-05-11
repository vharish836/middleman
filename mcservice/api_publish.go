package mcservice

import (
	"log"

	"github.com/vharish836/middleman/encdec"
	"github.com/vharish836/middleman/handler"
)

// Publish ...
func (s *Service) Publish(req *handler.JSONRequest) (*handler.JSONResponse, error) {
	if len(req.Params) < 3 {
		return &handler.JSONResponse{Error: map[string]interface{}{
			"code":  -1,
			"error": "Need exactly three arguments, refer to help",
		}}, nil
	}
	_, ok := req.Params[0].(string)
	if ok != true {
		return &handler.JSONResponse{Error: map[string]interface{}{
			"code":  -1,
			"error": "Invalid stream-identifier",
		}}, nil
	}
	entty, ok := req.Params[1].(string)
	if ok != true {
		return &handler.JSONResponse{Error: map[string]interface{}{
			"code":  -1,
			"error": "Invalid key",
		}}, nil
	}
	data, ok := req.Params[2].(string)
	if ok != true {
		return &handler.JSONResponse{Error: map[string]interface{}{
			"code":  -1,
			"error": "Invalid data",
		}}, nil
	}
	var key []byte
	k, ok := s.entityKeys.Get(entty)
	if ok != true {
		return &handler.JSONResponse{Error: map[string]interface{}{
			"code":  -1,
			"error": "Invalid key",
		}}, nil
	}
	key = k.([]byte)
	hexstr, err := encdec.EncryptData([]byte(data), key, s.cfg.Crypto.CryptoMode)
	if err != nil {
		log.Printf("could not encode: %s", err)
		return &handler.JSONResponse{Error: map[string]interface{}{
			"code":  -1,
			"error": "Internal Server Error",
		}}, nil
	}
	req.Params[2] = hexstr
	return s.platformAPI(req)
}
