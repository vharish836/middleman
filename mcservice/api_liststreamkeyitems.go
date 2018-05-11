package mcservice

import (
	"encoding/hex"
	"log"

	"github.com/vharish836/middleman/encdec"
	"github.com/vharish836/middleman/handler"
)

// Liststreamkeyitems ...
func (s *Service) Liststreamkeyitems(req *handler.JSONRequest) (*handler.JSONResponse, error) {
	if len(req.Params) < 2 {
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
	k, ok := s.entityKeys.Load(entty)
	if ok != true {
		return &handler.JSONResponse{Error: map[string]interface{}{
			"code":  -1,
			"error": "Invalid key",
		}}, nil
	}
	key := []byte(k.(string))
	rsp, err := s.PlatformAPI(req)
	if rsp.Result != nil {
		items, ok := rsp.Result.([]interface{})
		if ok != true {
			return &handler.JSONResponse{Error: map[string]interface{}{
				"code":  -1,
				"error": "Internal Server Error",
			}}, nil
		}
		for i := range items {
			item, ok := items[i].(map[string]interface{})
			if ok != true {
				return &handler.JSONResponse{Error: map[string]interface{}{
					"code":  -1,
					"error": "Internal Server Error",
				}}, nil
			}
			ciphertext, err := hex.DecodeString(item["data"].(string))
			if err != nil {
				log.Printf("could not decode hex: %s", err)
				return &handler.JSONResponse{Error: map[string]interface{}{
					"code":  -1,
					"error": "Internal Server Error",
				}}, nil
			}
			plaintext,err := encdec.DecryptData(ciphertext,key,s.cfg.CryptoMode)			
			if err != nil {
				return &handler.JSONResponse{Error: map[string]interface{}{
					"code":  -1,
					"error": "Internal Server Error",
				}}, nil
			}
			item["data"] = plaintext
		}
	}
	return rsp, err
}
