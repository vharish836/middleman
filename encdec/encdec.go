package encdec

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"log"
)

// CryptoMode ...
const (
	Aes256Gcm = iota
	Aes256Cbc
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
	return hexstr, nil
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
	return hexstr, nil
}

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

// EncryptData ...
func EncryptData(plaindata []byte, key []byte, mode int) (cipherdata string, err error ) {
	if mode == Aes256Gcm {
		cipherdata, err = aesEncryptGCM(plaindata, key)
	} else {
		cipherdata, err = aesEncryptCBC(plaindata, key)
	}
	return cipherdata,err
}

// DecryptData ...
func DecryptData(cipherdata []byte, key []byte, mode int) (plaindata string, err error) {
	if mode == Aes256Gcm {
		plaindata, err = aesDecryptGCM(cipherdata, key)
	} else {
		plaindata, err = aesDecryptCBC(cipherdata, key)
	}
	return plaindata,err
}