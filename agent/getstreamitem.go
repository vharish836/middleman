package agent

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"log"
)

// aesDecryptCBC ...
func aesDecryptCBC(ciphertext []byte, key []byte) (plaintext string, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	if len(ciphertext) < aes.BlockSize {
		log.Printf("data too short")
		return "", err
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	if len(ciphertext)%aes.BlockSize != 0 {
		log.Printf("data of unexpected length: expected %d longer", len(ciphertext)%aes.BlockSize)
		return "", err
	}
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)
	// remove PKCS5 padding
	padlen := int(ciphertext[len(ciphertext)-1])
	padding := ciphertext[len(ciphertext)-padlen:]
	for i := range padding {
		if padding[i] != byte(padlen) {
			log.Printf("unexpected padding")
			return "", err
		}
	}
	return string(ciphertext[:len(ciphertext)-padlen]), nil
}

// aesDecryptGCM ...
func aesDecryptGCM(ciphertext []byte, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Printf("could not create block: %s", err)
		return "", err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Printf("could not create GCM: %s", err)
		return "", err
	}
	nonce := ciphertext[len(ciphertext)-aesgcm.NonceSize():]
	ciphertext = ciphertext[:len(ciphertext)-aesgcm.NonceSize()]
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		log.Printf("could not open: %s", err)
		return "", err
	}
	return string(plaintext), nil
}

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
			plaintext,err = aesDecryptGCM(ciphertext,s.aeskey)
		} else {
			plaintext,err = aesDecryptCBC(ciphertext,s.aeskey)
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
