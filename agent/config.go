package agent

import (
	"errors"
	"fmt"

	"github.com/mitchellh/go-homedir"

	"github.com/BurntSushi/toml"
	"github.com/robfig/config"
)

// MultiChainConfig ...
type MultiChainConfig struct {
	ChainName   string
	RPCPort     int
	RPCUser     string
	RPCPassword string
}

// Config ...
type Config struct {
	UserName   string
	PassWord   string
	EasyAPI    bool
	MultiChain MultiChainConfig
}

// LoadPrimaryConfig ...
func LoadPrimaryConfig(file string) (*Config, error) {
	cfg := &Config{EasyAPI: true}
	cfile, ferr := homedir.Expand(file)
	if ferr != nil {
		return nil, ferr
	}
	_, err := toml.DecodeFile(cfile, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// LoadSecondaryConfig ...
func LoadSecondaryConfig(cfg *Config) (err error) {
	if cfg.MultiChain.ChainName != "" {
		mconffile := "~/.multichain/" + cfg.MultiChain.ChainName + "/multichain.conf"
		mfile, ferr := homedir.Expand(mconffile)
		if ferr != nil {
			return ferr
		}
		c, cerr := config.ReadDefault(mfile)
		if cerr != nil {
			return cerr
		}
		cfg.MultiChain.RPCUser, err = c.RawStringDefault("rpcuser")
		if err != nil {
			return err
		}
		cfg.MultiChain.RPCPassword, err = c.RawStringDefault("rpcpassword")
		if err != nil {
			return err
		}
		if cfg.MultiChain.RPCPort == 0 {
			cfg.MultiChain.RPCPort, err = c.Int("DEFAULT","rpcport")
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// GetConfig ...
func GetConfig(file string) (*Config, error) {
	cfg, err := LoadPrimaryConfig(file)
	if err != nil {
		return nil, err
	}
	err = CheckConfig(cfg)
	// this means all requierd parameters already present in primary config
	if err == nil {
		return cfg, nil
	}
	err = LoadSecondaryConfig(cfg)
	if err != nil {
		return nil, err
	}
	err = CheckConfig(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// CheckConfig ...
func CheckConfig(c *Config) error {
	if c.UserName == "" {
		return errors.New("Missing required parameter: username")
	}
	if c.PassWord == "" {
		return errors.New("Missing required parameter: password")
	}
	if c.MultiChain == (MultiChainConfig{}) {
		return errors.New("Missing required table: multichain")
	}
	if c.MultiChain.ChainName == "" {
		return errors.New("Missing required parameter: chainname under multichain table")
	}
	if c.MultiChain.RPCPort == 0 {
		serr := fmt.Sprintf("Missing required parameter: rpcport in multichain.conf for chain %s",
			c.MultiChain.ChainName)
		return errors.New(serr)
	}
	if c.MultiChain.RPCUser == "" {
		serr := fmt.Sprintf("Missing required parameter: rpcuser in multichain.conf for chain %s",
			c.MultiChain.ChainName)
		return errors.New(serr)
	}
	if c.MultiChain.RPCPassword == "" {
		serr := fmt.Sprintf("Missing required parameter: rpcpassword in multichain.conf for chain %s",
			c.MultiChain.ChainName)
		return errors.New(serr)
	}
	return nil
}
