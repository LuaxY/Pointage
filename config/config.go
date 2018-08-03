package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"
)

const DebugMode = false
const UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Safari/537.36"

var cfg = LoadConfiguration()

type Config struct {
	ADP struct {
		Username string   `json:"username"`
		Password string   `json:"password"`
		Auto     []string `json:"auto"`
	} `json:"adp"`

	Friday struct {
		Shutdown bool   `json:"shutdown"`
		At       string `json:"at"`
	} `json:"friday"`
}

func LoadConfiguration() Config {
	var config Config

	file := flag.String("c", "config.json", "config file")
	flag.Parse()

	log.Printf("Read config file: %s", *file)

	configFile, err := os.Open(*file)
	defer configFile.Close()

	if err != nil {
		log.Print(err)
	}

	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&config)

	if err != nil {
		log.Print(err)
	}

	return config
}

func Get() *Config {
	return &cfg
}
