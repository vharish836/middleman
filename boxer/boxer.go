package boxer

import (
	"encoding/hex"
	"encoding/json"
	"sync"

	cache "github.com/patrickmn/go-cache"
	"github.com/vharish836/middleman/cipher"
)

// KeyInfo ...
type KeyInfo struct {
	ID      string
	Value   string
	Entity  string
	Default bool
}

// CryptoConfig ...
type CryptoConfig struct {
	Mode int
	Keys []KeyInfo
}

// CacheConfig ...
type CacheConfig struct {
	TTL             string
	CleanupInterval string
}

// Config ...
type Config struct {
	KeyCache CacheConfig
	Crypto   CryptoConfig
}

// Boxer ...
type Boxer struct {
	cfg         *Config
	entityCache *cache.Cache
	defaultKeys sync.Map
}

// NewBoxer ...
func NewBoxer(cfg *Config) (*Boxer, error) {
	ch := cache.New(cache.NoExpiration, cache.NoExpiration)
	b := &Boxer{
		cfg:         cfg,
		entityCache: ch,
	}
	err := b.loadEntityMap()
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Box encrypts the given data and boxes it into Block along with
// header containing all relevant information needed to UnBox later
func (b *Boxer) Box(data []byte, entityID string) (string, error) {
	k, ok := b.defaultKeys.Load(entityID)
	if ok != true {
		return "", errNoKeys
	}
	key := k.(KeyInfo)
	cipherdata, salt, err := cipher.EncryptData(data, []byte(key.Value), b.cfg.Crypto.Mode)
	if err != nil {
		return "", err
	}
	hdr := headerV1{
		Algorithm: b.cfg.Crypto.Mode,
		KeyID:     key.ID,
		EntityID:  entityID,
		Salt:      hex.EncodeToString(salt),
	}
	box := boxV1{
		Version: HeaderVersion1,
		Header:  hdr,
		Data:    hex.EncodeToString(cipherdata),
	}
	bbuf, err := json.Marshal(box)
	if err != nil {
		return "", err
	}
	hexstr := hex.EncodeToString(bbuf)
	return hexstr, nil
}

// UnBox decrypts the given Block using the header information
// and provides the data
func (b *Boxer) UnBox(block string) (string, error) {
	buf, err := hex.DecodeString(block)
	if err != nil {
		return "", err
	}
	bx := boxV1{}
	err = json.Unmarshal(buf, &bx)
	if err != nil {
		return "", err
	}
	kc, ok := b.entityCache.Get(bx.Header.EntityID)
	if ok != true {
		return "", errNoKeys
	}
	keyCache := kc.(*cache.Cache)
	k, ok := keyCache.Get(bx.Header.KeyID)
	if ok != true {
		return "", errNoKeys
	}
	key := k.(string)
	salt, err := hex.DecodeString(bx.Header.Salt)
	if err != nil {
		return "", err
	}
	cipherdata, err := hex.DecodeString(bx.Data)
	if err != nil {
		return "", err
	}
	return cipher.DecryptData(cipherdata, []byte(key),
		salt, bx.Header.Algorithm)
}
