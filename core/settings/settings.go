package settings

import (
	"fmt"
	"os"
	"time"
)

type Settings struct {
	HTTP HTTPConfig `yaml:"http"`
	MongoDB      MongoDBConfig      `yaml:"mongodb"`
	EtherscanAPI EtherscanAPIConfig `yaml:"etherscan-api"`
}

type HTTPConfig struct {
	Host         string `yaml:"host"`
	Port         uint16 `yaml:"port"`
}

func (h HTTPConfig) URL() string {
	// return fmt.Sprint(h.Host, ":", h.Port)
	herokuPORT := os.Getenv("PORT")
	return ":" + herokuPORT
}

type MongoDBConfig struct {
	DatabaseName string `yaml:"name"`
	Host         string `yaml:"host"`
	Port         uint16 `yaml:"port"`
}
type EtherscanAPIConfig struct {
	URL      string        `yaml:"url"`
	ReqDelay time.Duration `yaml:"req-delay"`
	Key      string        `yaml:"key"`
}

func (c MongoDBConfig) ConnectionURL() string {
	// return fmt.Sprint("mongodb://", c.Host, ":", c.Port)
	return os.Getenv("MONGODB_URI")
}

func Init() (settings *Settings, err error) {
	settings = &Settings{EtherscanAPI: EtherscanAPIConfig{
		URL:      "https://api.etherscan.io/api",
		ReqDelay: time.Second,
		Key:      os.Getenv("API_KEY"),
	},
	MongoDB: MongoDBConfig{
		DatabaseName: "mongo",
	},
	}
	// path := flag.String("config", "conf.yaml", "path to configuration file (must end with conf.yaml)")
	// flag.Parse()
	//
	// data, err := os.ReadFile(*path)
	// if err != nil {
	// 	return
	// }
	//
	// settings = &Settings{}

	// err = yaml.Unmarshal(data, settings)
	// if err != nil {
	// 	return
	// }
	//
	// err = validate(settings)
	// if err != nil {
	// 	return nil, errors.Wrap(err, "config validation error")
	// }
	return
}

func validate(settings *Settings) (err error) {
	// HTTP
	if settings.HTTP.Host == "" ||
		settings.HTTP.Port == 0 {
		err = fmt.Errorf("empty HTTP config")
		return
	}

	// mongoDB
	if settings.MongoDB.Host == "" ||
		settings.MongoDB.Port == 0 ||
		settings.MongoDB.DatabaseName == "" {
		err = fmt.Errorf("empty mongoDB config")
		return
	}

	// Etherscan API
	if settings.EtherscanAPI.URL == "" ||
		settings.EtherscanAPI.Key == "" {
		err = fmt.Errorf("empty etherscan api config")
		return
	}
	if settings.EtherscanAPI.ReqDelay < 500*time.Millisecond {
		err = fmt.Errorf("delay between etherscan api request must be more then 0.5 sec")
		return
	}
	return nil
}
