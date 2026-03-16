package config

import (
	"os"
	"gopkg.in/yaml.v3"
)

// Configurazione di un singolo nodo worker
type NodeConfig struct {
	URL  string `yaml:"url"`
	Zone string `yaml:"zone"`
}

// Intero file config.yaml
type Config struct {
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
	
	Balancer struct {
		Algorithm string `yaml:"algorithm"`
		APIKey    string `yaml:"api_key"`
	} `yaml:"balancer"`
	
	Nodes []NodeConfig `yaml:"nodes"`
}

// Legge il file yaml e lo trasforma in una struct Config
func LoadConfig(filename string) (*Config, error) {
	// Lettura file da disco
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Variabile vuota di tipo Config
	var cfg Config

	// Unmarshal dei dati YAML dentro la variabile
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}