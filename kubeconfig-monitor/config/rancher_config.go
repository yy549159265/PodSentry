package config

import (
	"os"
)

type Config struct {
	RancherServer string
	RancherToken  string
	PodNamespace  string
	PodName       string
}

func LoadConfig() *Config {
	return &Config{
		RancherServer: os.Getenv("RANCHER_SERVER"),
		RancherToken:  os.Getenv("RANCHER_TOKEN"),
		PodNamespace:  os.Getenv("POD_NAMESPACE"),
		PodName:       os.Getenv("POD_NAME"),
	}
}
