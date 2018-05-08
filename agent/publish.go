package agent

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"log"
)

// aesEncryptCBC ...
func aesEncryptCBC(plaindata []byte, key []byte) (hexstr string, err error) {
	data := plaindata
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	padlen := block.BlockSize() - (len(data) % block.BlockSize())
	ciphertext := make([]byte, aes.BlockSize+padlen+len(data))
	// PKCS5 padding
	if padlen == 0 {
		padlen = aes.BlockSize
	}
	if padlen != 0 {
		padding := make([]byte, padlen)
		for i := range padding {
			padding[i] = byte(padlen)
		}
		data = append(data, padding...)
	}
	
	iv := ciphertext[:aes.BlockSize]
	_, err = rand.Read(iv)
	if err != nil {
		return "", err
	}
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], data)
	hexstr = hex.EncodeToString(ciphertext)
	return hexstr,nil
}

// aesEncryptGCM ...
func aesEncryptGCM(plaindata []byte, key []byte) (hexstr string, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Printf("could not create GCM: %s", err)
		return "", err
	}
	nonce := make([]byte, aesgcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		log.Printf("could not init nounce: %s", err)
		return "", err
	}
	ciphertext := aesgcm.Seal(nil, nonce, plaindata, nil)
	ciphertext = append(ciphertext, nonce...)
	hexstr = hex.EncodeToString(ciphertext)
	return hexstr,nil
}

// Publish ...
func (s *Service) Publish(req *JSONRequest) (*JSONResponse, error) {
	if len(req.Params) < 3 {
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
			"error": "Invalid key",
		}}, nil
	}
	data, ok := req.Params[2].(string)
	if ok != true {
		return &JSONResponse{Error: map[string]interface{}{
			"code":  -1,
			"error": "Invalid data",
		}}, nil
	}
	var hexstr string
	var err error
	if s.cfg.AESMode == GCMMode {
		hexstr,err = aesEncryptGCM([]byte(data),s.aeskey)
	} else {
		hexstr,err = aesEncryptCBC([]byte(data),s.aeskey)
	}
	
	if err != nil {
		return &JSONResponse{Error: map[string]interface{}{
			"code":  -1,
			"error": "Internal Server Error",
		}}, nil
	}
	req.Params[2] = hexstr
	return s.PlatformAPI(req)
}
