package config

import (
	"fmt"
	"github.com/baidubce/bce-sdk-go/services/cdn/api"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	Acme    AcmeInfo     `yaml:"acme"`
	Domains []DomainInfo `yaml:"domains"`
}
type AcmeInfo struct {
	Email string `yaml:"email"`
}

type DomainInfo struct {
	Domain   string `yaml:"domain"`
	Provider string `yaml:"provider"`
	Baidu    *Baidu `yaml:"baidu"`
}

type Baidu struct {
	Origin []api.OriginPeer `yaml:"origin"`
	Form   string           `yaml:"form"`
	Dsa    *api.DSAConfig   `yaml:"dsa"`
	Cname  bool             `yaml:"cname"`
}

// LoadConfig 读取配置文件
func LoadConfig(filename string) (Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return Config{}, fmt.Errorf("无法读取配置文件: %v", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return Config{}, fmt.Errorf("解析配置文件失败: %v", err)
	}
	return config, nil
}

// SaveConfig 保存配置文件
func SaveConfig(filename string, config Config) error {
	data, err := yaml.Marshal(&config)
	if err != nil {
		return fmt.Errorf("无法序列化配置文件: %v", err)
	}

	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	return nil
}
