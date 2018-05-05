package agent

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"log"
)

// GetStreamItem ...
func (s *Service) GetStreamItem(req *JSONRequest) (*JSONResponse, error) {
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
			return rsp, err
		}
		ciphertext, err := hex.DecodeString(rdata["data"].(string))
		if err != nil {
			log.Printf("could not decode hex: %s", err)
		}
		block, err := aes.NewCipher(s.aeskey)
		if err != nil {
			return nil, err
		}
		if len(ciphertext) < aes.BlockSize {
			log.Printf("data too short")
			return rsp, err
		}
		iv := ciphertext[:aes.BlockSize]
		ciphertext = ciphertext[aes.BlockSize:]
		if len(ciphertext)%aes.BlockSize != 0 {
			log.Printf("data of unexpected length: expected %d longer", len(ciphertext)%aes.BlockSize)
		}
		mode := cipher.NewCBCDecrypter(block, iv)
		mode.CryptBlocks(ciphertext, ciphertext)
		// remove PKCS5 padding
		padlen := int(ciphertext[len(ciphertext)-1])
		padding := ciphertext[len(ciphertext)-padlen:]
		log.Print(padding)
		for i := range padding {
			if padding[i] != byte(padlen) {
				log.Printf("unexpected padding")
				return rsp, err
			}
		}
		rdata["data"] = string(ciphertext[:len(ciphertext)-padlen])
	}
	return rsp, err
}

// GetStreamItemGCM ...
func (s *Service) GetStreamItemGCM(req *JSONRequest) (*JSONResponse, error) {
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
	req.Method = "getstreamitem"
	rsp, err := s.PlatformAPI(req)
	if rsp.Result != nil {
		rdata, ok := rsp.Result.(map[string]interface{})
		if ok != true {
			log.Printf("could not type assert to object")
			return rsp, err
		}
		ciphertext, err := hex.DecodeString(rdata["data"].(string))
		if err != nil {
			log.Printf("could not decode hex: %s", err)
			return rsp,err
		}
		block, err := aes.NewCipher(s.aeskey)
		if err != nil {
			log.Printf("could not create block: %s",err)
			return rsp, err
		}
		aesgcm,err := cipher.NewGCM(block)
		if err != nil {
			log.Printf("could not create GCM: %s",err)
			return rsp,err
		}
		nonce := ciphertext[len(ciphertext)-aesgcm.NonceSize():]
		ciphertext = ciphertext[:len(ciphertext)-aesgcm.NonceSize()]
		plaintext,err := aesgcm.Open(nil,nonce,ciphertext,nil)
		if err != nil {
			log.Printf("could not open: %s",err)
			return rsp,err
		}
		rdata["data"] = string(plaintext)
	}
	return rsp, err
}