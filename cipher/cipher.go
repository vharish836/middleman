package cipher

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"log"
)

// CryptoMode ...
const (
	Aes256Gcm = iota
	Aes256Cbc
)

// aesEncryptCBC ...
func aesEncryptCBC(plaindata []byte, key []byte) (cipherdata []byte, iv []byte, err error) {
	data := plaindata
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}
	padlen := block.BlockSize() - (len(data) % block.BlockSize())
	cipherdata = make([]byte, padlen+len(data))
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

	iv = make([]byte, aes.BlockSize)
	_, err = rand.Read(iv)
	if err != nil {
		return nil, nil, err
	}
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherdata, data)
	return cipherdata, iv, nil
}

// aesEncryptGCM ...
func aesEncryptGCM(plaindata []byte, key []byte) (cipherdata []byte, nonce []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Printf("could not create GCM: %s", err)
		return nil, nil, err
	}
	nonce = make([]byte, aesgcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		log.Printf("could not init nounce: %s", err)
		return nil, nil, err
	}
	cipherdata = aesgcm.Seal(nil, nonce, plaindata, nil)
	return cipherdata, nonce, nil
}

// aesDecryptCBC ...
func aesDecryptCBC(cipherdata []byte, iv []byte, key []byte) (plaintext string, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	if len(cipherdata) < aes.BlockSize {
		log.Printf("data too short")
		return "", err
	}
	if len(cipherdata)%aes.BlockSize != 0 {
		log.Printf("data of unexpected length: expected %d longer", len(cipherdata)%aes.BlockSize)
		return "", err
	}
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(cipherdata, cipherdata)
	// remove PKCS5 padding
	padlen := int(cipherdata[len(cipherdata)-1])
	padding := cipherdata[len(cipherdata)-padlen:]
	for i := range padding {
		if padding[i] != byte(padlen) {
			log.Printf("unexpected padding")
			return "", err
		}
	}
	return string(cipherdata[:len(cipherdata)-padlen]), nil
}

// aesDecryptGCM ...
func aesDecryptGCM(cipherdata []byte, nonce []byte, key []byte) (string, error) {
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
	plaintext, err := aesgcm.Open(nil, nonce, cipherdata, nil)
	if err != nil {
		log.Printf("could not open: %s", err)
		return "", err
	}
	return string(plaintext), nil
}

// EncryptData ...
func EncryptData(plaindata []byte, key []byte, mode int) ([]byte, []byte, error) {
	if mode == Aes256Gcm {
		return aesEncryptGCM(plaindata, key)
	}
	return aesEncryptCBC(plaindata, key)

}

// DecryptData ...
func DecryptData(cipherdata []byte, key []byte, salt []byte, mode int) (string, error) {
	if mode == Aes256Gcm {
		return aesDecryptGCM(cipherdata, salt, key)
	} 
	return aesDecryptCBC(cipherdata, salt, key)	
}
