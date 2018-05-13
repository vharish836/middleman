package boxer

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"sync"
	"time"

	cache "github.com/patrickmn/go-cache"
	"github.com/vharish836/middleman/cipher"
)

// HeaderVersion ...
const (
	HeaderVersion1 = 1
)

type headerV1 struct {
	Algorithm int    `json:"algorithm"`
	KeyID     string `json:"keyid"`
	EntityID  string `json:"entityid"`
	Salt      string `json:"salt"`
}

type boxV1 struct {
	Version int      `json:"version"`
	Header  headerV1 `json:"blockheader"`
	Data    string   `json:"data"`
}

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

var errNoKeys = errors.New("no keys configured")
var errInvalidConfig = errors.New("invalid config")
var errUnknownVersion = errors.New("unknown header version")

func (b *Boxer) loadEntityMap() error {
	t, err := time.ParseDuration(b.cfg.KeyCache.TTL)
	if err != nil {
		return errInvalidConfig
	}
	c, err := time.ParseDuration(b.cfg.KeyCache.CleanupInterval)
	if err != nil {
		return errInvalidConfig
	}
	if b.cfg.Crypto.Keys == nil {
		return errNoKeys
	}

	keys := b.cfg.Crypto.Keys
	for i := range keys {
		entity := keys[i].Entity
		keyID := keys[i].ID
		keyValue, err := hex.DecodeString(keys[i].Value)
		if err != nil {
			return err
		}
		kc, ok := b.entityCache.Get(entity)
		if ok != true {
			kc = cache.New(t, c)
			b.entityCache.SetDefault(entity, kc)
		}
		keyCache := kc.(*cache.Cache)
		keyCache.SetDefault(keyID, string(keyValue))
		if keys[i].Default {
			if _, ok := b.defaultKeys.Load(entity); ok == true {
				return errInvalidConfig
			}
			keys[i].Value = string(keyValue)
			b.defaultKeys.Store(entity, keys[i])
		}
	}
	return nil
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
