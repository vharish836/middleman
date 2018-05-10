package agent

import (
	"encoding/hex"
	"log"
)

// Getstreamitem ...
func (s *Service) Getstreamitem(req *JSONRequest) (*JSONResponse, error) {
	if len(req.Params) < 2 {
		return &JSONResponse{Error: map[string]interface{}{
			"code":  -1,
			"error": "Need exactly three arguments, refer to help",
		}}, nil
	}
	_, ok := req.Params[0].(string)
	if ok != true {
		return &JSONResponse{Error: map[string]interface{}{
			"code":  -1,
			"error": "Invalid stream-identifier",
		}}, nil
	}
	_, ok = req.Params[1].(string)
	if ok != true {
		return &JSONResponse{Error: map[string]interface{}{
			"code":  -1,
			"error": "Invalid tx-id",
		}}, nil
	}
	rsp, err := s.PlatformAPI(req)
	if rsp.Result != nil {
		rdata, ok := rsp.Result.(map[string]interface{})
		if ok != true {
			log.Printf("could not type assert to object")
			return &JSONResponse{Error: map[string]interface{}{
				"code":  -1,
				"error": "Internal Server Error",
			}}, nil
		}
		ciphertext, err := hex.DecodeString(rdata["data"].(string))
		if err != nil {
			log.Printf("could not decode hex: %s", err)
			return &JSONResponse{Error: map[string]interface{}{
				"code":  -1,
				"error": "Internal Server Error",
			}}, nil
		}
		var plaintext string
		if s.cfg.AESMode == GCMMode {
			plaintext, err = aesDecryptGCM(ciphertext, s.nativeKey)
		} else {
			plaintext, err = aesDecryptCBC(ciphertext, s.nativeKey)
		}
		if err != nil {
			return &JSONResponse{Error: map[string]interface{}{
				"code":  -1,
				"error": "Internal Server Error",
			}}, nil
		}

		rdata["data"] = plaintext
	}
	return rsp, err
}
