package pkg

import (
	"encoding/json"
	"os"
)

// raw ConfigServer data type for json
type ConfigServer struct {
	Version  string `json:"version"`
	Listener struct {
		BackendApi struct {
			Address string `json:"address"`
			Port    int32  `json:"port"`
		} `json:"backend_api"`
	} `json:"listener"`
	Database struct {
		PostgreSQL struct {
			Main struct {
				Host     string `json:"host"`
				Port     int32  `json:"port"`
				User     string `json:"user"`
				Password string `json:"password"`
				Database string `json:"database"`
				SslMode  string `json:"sslmode"`
			} `json:"main"`
		} `json:"postgresql"`
		Redis struct {
			Main struct {
				Host     string `json:"host"`
				Port     int32  `json:"port"`
				User     string `json:"user"`
				Password string `json:"password"`
				Db       int32  `json:"db"`
			} `json:"main"`
		} `json:"redis"`
	} `json:"database"`
	Security struct {
		WhitelistOrigin []string `json:"whitelist_origin"`
		WhitelistHost   []string `json:"whitelist_host"`
		BlockCipher     struct {
			Default struct {
				Iv string `json:"iv"`
				Ik string `json:"ik"`
			} `json:"default"`
		} `json:"block_cipher"`
	} `json:"security"`
}

// load config file from file path
//
// params:
// 	fp string - filepath, relative from the executeable
//
// return: (ConfigServer, error)
func ConfigServerLoad(fp string) (ConfigServer, error) {
	var cfg ConfigServer

	content, err := os.ReadFile(fp)
	if err != nil {
		return cfg, err
	}

	err = json.Unmarshal(content, &cfg)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}
