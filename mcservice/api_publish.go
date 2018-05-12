package mcservice

import (
	"log"

	"github.com/vharish836/middleman/cipher"
)

func (s *MCService) publish(req *JSONRequest) (*JSONResponse, error) {
	if len(req.Params) < 3 {
		return nil, errNumParameter
	}
	_, ok := req.Params[0].(string)
	if ok != true {
		return nil, errParameter
	}
	_, ok = req.Params[1].(string)
	if ok != true {
		return nil, errParameter
	}
	data, ok := req.Params[2].(string)
	if ok != true {
		return nil, errParameter
	}
	var key []byte
	k, ok := s.entityKeys.Get(s.nativeEntity)
	if ok != true {
		return nil, errParameter
	}
	key = k.([]byte)
	hexstr, err := cipher.EncryptData([]byte(data), key, s.cfg.Crypto.CryptoMode)
	if err != nil {
		log.Printf("could not encode: %s", err)
		return nil, errInternal
	}
	req.Params[2] = hexstr
	return s.platformAPI(req)
}
