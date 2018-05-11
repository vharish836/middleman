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
		return nil,errNumParameter
	}
	_, ok := req.Params[0].(string)
	if ok != true {
		return nil,errParameter
	}
	_, ok = req.Params[1].(string)
	if ok != true {
		return nil,errParameter
	}
	rsp, err := s.platformAPI(req)
	if rsp.Result != nil {
		rdata, ok := rsp.Result.(map[string]interface{})
		if ok != true {
			log.Printf("could not type assert to object")
			return nil,errInternal
		}
		ciphertext, err := hex.DecodeString(rdata["data"].(string))
		if err != nil {
			log.Printf("could not decode hex: %s", err)
			return nil,errInternal
		}
		k, ok := s.entityKeys.Get(s.nativeEntity)
		if ok != true {
			log.Printf("could not get native key")
			return nil,errInternal
		}
		key := k.([]byte)
		plaintext, err := encdec.DecryptData(ciphertext, key, s.cfg.Crypto.CryptoMode)
		if err != nil {
			log.Printf("could not decrypt: %s",err)
			return nil,errInternal
		}

		rdata["data"] = plaintext
	}
	return rsp, err
}
