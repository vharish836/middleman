package config

import (
	"github.com/mitchellh/go-homedir"

	"github.com/BurntSushi/toml"
	"github.com/robfig/config"
	"github.com/vharish836/middleman/boxer"
	"github.com/vharish836/middleman/mcservice"
)

// Config ...
type Config struct {
	Boxer      boxer.Config
	MultiChain mcservice.Config
}

// loadPrimaryConfig ...
func loadPrimaryConfig(file string) (*Config, error) {
	cfg := &Config{}
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

// loadSecondaryConfig ...
func loadSecondaryConfig(cfg *Config) (err error) {
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
			cfg.MultiChain.RPCPort, err = c.Int("DEFAULT", "rpcport")
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// GetConfig ...
func GetConfig(file string) (*Config, error) {
	cfg, err := loadPrimaryConfig(file)
	if err != nil {
		return nil, err
	}
	err = loadSecondaryConfig(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
