package agent

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"log"
)

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
	block, err := aes.NewCipher(s.aeskey)
	if err != nil {
		return nil, err
	}
	plaindata := []byte(data)
	padlen := len(plaindata) % block.BlockSize()
	ciphertext := make([]byte, aes.BlockSize+padlen+len(plaindata))
	// PKCS5 padding
	if padlen == 0 {
		padlen = aes.BlockSize
	}
	if padlen != 0 {
		padding := make([]byte, padlen)
		for i := range padding {
			padding[i] = byte(padlen)
		}
		plaindata = append(plaindata, padding...)
	}
	iv := ciphertext[:aes.BlockSize]
	_, err = rand.Read(iv)
	if err != nil {
		return nil, err
	}
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaindata)
	hexstr := hex.EncodeToString(ciphertext)
	req.Params[2] = hexstr
	return s.PlatformAPI(req)
}

// PublishGCM ...
func (s *Service) PublishGCM(req *JSONRequest) (*JSONResponse, error) {
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
	plaindata := []byte(data)
	block, err := aes.NewCipher(s.aeskey)
	if err != nil {
		log.Printf("could not craete cipher: %s", err)
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Printf("could not create GCM: %s", err)
		return nil, err
	}
	nonce := make([]byte, aesgcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		log.Printf("could not init nounce: %s", err)
		return nil, err
	}
	ciphertext := aesgcm.Seal(nil, nonce, plaindata, nil)
	ciphertext = append(ciphertext, nonce...)
	hexstr := hex.EncodeToString(ciphertext)
	req.Method = "publish"
	req.Params[2] = hexstr
	return s.PlatformAPI(req)
}
