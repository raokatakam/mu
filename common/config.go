package common

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"fmt"
	"log"
)

// Config defines the structure of the yml file for the mu config
type Config struct {
	Environments []Environment `yaml:"environments,omitempty"`
	Service Service `yaml:"service,omitempty"`
}

// Environment defines the env that will be created
type Environment struct {
	Name string `yaml:"name"`
	Loadbalancer EnvironmentLoadBalancer `yaml:"loadbalancer,omitempty"`
	Cluster EnvironmentCluster `yaml:"cluster,omitempty"`
}

// EnvironmentLoadBalancer defines the env load balancer that will be created
type EnvironmentLoadBalancer struct {
	Hostname string `yaml:"hostname,omitempty"`
}

// EnvironmentCluster defines the env cluster that will be created
type EnvironmentCluster struct {
	DesiredCapacity int `yaml:"desiredCapacity,omitempty"`
	MaxSize int `yaml:"maxSize,omitempty"`
}

// Service defines the service that will be created
type Service struct {
	DesiredCount int `yaml:"desiredCount,omitempty"`
	Pipeline ServicePipeline `yaml:"pipeline,omitempty"`
}

// ServicePipeline defines the service pipeline that will be created
type ServicePipeline struct {
}

// NewConfig create a new config object
func NewConfig() *Config {
	return &Config{}
}

// LoadConfig loads config object from local file
func LoadConfig(config *Config, configFile string) {
	yamlConfig, err := ioutil.ReadFile( configFile )
	if err != nil {
		fmt.Printf("WARN: Unable to find config file: %v\n", err)
	} else {
		loadYamlConfig(config, yamlConfig)
	}

}

func loadYamlConfig(config *Config, yamlConfig []byte)  *Config {
	err := yaml.Unmarshal(yamlConfig, config)
	if err != nil {
		log.Panicf("Invalid config file: %v", err)
	}

	return config
}
