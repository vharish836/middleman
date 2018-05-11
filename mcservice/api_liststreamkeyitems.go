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
	k, ok := s.entityKeys.Get(s.nativeEntity)
	if ok != true {
		return nil,errParameter
	}
	key := k.([]byte)
	rsp, err := s.platformAPI(req)
	if rsp.Result != nil {
		items, ok := rsp.Result.([]interface{})
		if ok != true {
			log.Printf("unexpected result: %+v",rsp.Result)
			return nil,errInternal
		}
		for i := range items {
			item, ok := items[i].(map[string]interface{})
			if ok != true {
				log.Printf("unexpected result: %+v",items[i])
				return nil,errInternal
			}
			ciphertext, err := hex.DecodeString(item["data"].(string))
			if err != nil {
				log.Printf("could not decode hex: %s", err)
				return nil,errInternal
			}
			plaintext, err := encdec.DecryptData(ciphertext, key, s.cfg.Crypto.CryptoMode)
			if err != nil {
				log.Printf("could not decrypt: %s",err)
				return nil,errInternal
			}
			item["data"] = plaintext
		}
	}
	return rsp, err
}
