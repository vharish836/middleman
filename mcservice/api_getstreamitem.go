package mcservice

import (
	"encoding/hex"
	"log"

	"github.com/vharish836/middleman/encdec"
	"github.com/vharish836/middleman/handler"
)

// Getstreamitem ...
func (s *Service) Getstreamitem(req *handler.JSONRequest) (*handler.JSONResponse, error) {
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
	_, ok = req.Params[1].(string)
	if ok != true {
		return &handler.JSONResponse{Error: map[string]interface{}{
			"code":  -1,
			"error": "Invalid tx-id",
		}}, nil
	}
	rsp, err := s.PlatformAPI(req)
	if rsp.Result != nil {
		rdata, ok := rsp.Result.(map[string]interface{})
		if ok != true {
			log.Printf("could not type assert to object")
			return &handler.JSONResponse{Error: map[string]interface{}{
				"code":  -1,
				"error": "Internal Server Error",
			}}, nil
		}
		ciphertext, err := hex.DecodeString(rdata["data"].(string))
		if err != nil {
			log.Printf("could not decode hex: %s", err)
			return &handler.JSONResponse{Error: map[string]interface{}{
				"code":  -1,
				"error": "Internal Server Error",
			}}, nil
		}
		plaintext,err := encdec.DecryptData(ciphertext,s.nativeKey,s.cfg.CryptoMode)		
		if err != nil {
			return &handler.JSONResponse{Error: map[string]interface{}{
				"code":  -1,
				"error": "Internal Server Error",
			}}, nil
		}

		rdata["data"] = plaintext
	}
	return rsp, err
}
