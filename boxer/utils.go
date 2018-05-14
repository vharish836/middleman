package boxer

import (
	"encoding/hex"
	"errors"
	"log"
	"time"

	cache "github.com/patrickmn/go-cache"
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

var errNoKeys = errors.New("no keys configured")
var errInvalidConfig = errors.New("invalid config")
var errUnknownVersion = errors.New("unknown header version")

func (b *Boxer) loadEntityMap() error {
	entities := map[string]bool{}
	t, err := time.ParseDuration(b.cfg.KeyCache.TTL)
	if err != nil {
		return err
	}
	c, err := time.ParseDuration(b.cfg.KeyCache.CleanupInterval)
	if err != nil {
		return err
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
		if len(keyValue) != 32 {
			log.Printf("configured key with id %s for entity %s has incorrect length",keyID,entity)
			return errInvalidConfig
		}
		kc, ok := b.entityCache.Get(entity)
		if ok != true {
			kc = cache.New(t, c)
			b.entityCache.SetDefault(entity, kc)
			entities[entity] = true
		}
		keyCache := kc.(*cache.Cache)
		keyCache.SetDefault(keyID, string(keyValue))
		if keys[i].Default {
			if _, ok := b.defaultKeys.Load(entity); ok == true {
				log.Printf("only one default key allowed per entity")
				return errInvalidConfig
			}
			keys[i].Value = string(keyValue)
			b.defaultKeys.Store(entity, keys[i])
			delete(entities, entity)
		}
	}
	for k := range entities {
		log.Printf("entity %s does not have a default key",k)
		return errInvalidConfig
	}
	return nil
}
