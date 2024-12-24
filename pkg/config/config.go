package config

import (
	"fmt"
	"github.com/baidubce/bce-sdk-go/services/cdn/api"
	"github.com/wjlin0/acmeGoBaidu/pkg/baidu/dns01"
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
	Domain   string    `yaml:"domain"`
	Provider string    `yaml:"provider"`
	To       string    `yaml:"to"`
	Baidu    *BaiduYun `yaml:"baidu,omitempty"`
	AliYun   *AliYun   `yaml:"ali,omitempty"`
}
type AliYun struct {
	Kodo *Kodo `yaml:"kodo"`
}
type Kodo struct {
	Bucket    string `yaml:"bucket"`
	Region    string `yaml:"region"`
	CnameInfo *Cname `yaml:"cname"`
}

type BaiduYun struct {
	CDN *BaiduYunCDN `yaml:"cdn"`
}

type BaiduYunCDN struct {
	Origin        []api.OriginPeer `yaml:"origin"`
	OriginTimeout *OriginTimeout   `yaml:"OriginTimeout"`
	Form          string           `yaml:"form"`
	Dsa           *api.DSAConfig   `yaml:"dsa"`
	CnameInfo     *Cname           `yaml:"cname"`
	IPv6          bool             `yaml:"ipv6"`
	QUIC          bool             `yaml:"quic"`
	HTTP2         bool             `yaml:"http2"`
	HTTP3         bool             `yaml:"http3"`
	Seo           *api.SeoSwitch   `yaml:"seo"`
}
type OriginTimeout struct {
	ConnectTimeout int `json:"connectTimeout,omitempty"`
	LoadTimeout    int `json:"loadTimeout,omitempty"`
}

type Cname struct {
	Enabled bool   `yaml:"enabled"`
	Value   string `yaml:"value"`
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
	// default value
	for i := range config.Domains {
		tmp := config.Domains[i]
		if tmp.Baidu == nil {
			continue
		}
		config.Domains[i].Domain = dns01.UnFqdn(tmp.Domain)
		if tmp.Baidu != nil {
			if tmp.Baidu.CDN != nil {
				if tmp.Baidu.CDN.CnameInfo.Enabled == true {
					config.Domains[i].Baidu.CDN.CnameInfo.Value = fmt.Sprintf("%s.a.bdydns.com.", config.Domains[i].Domain)
				}

				if tmp.Baidu.CDN.OriginTimeout != nil {
					if tmp.Baidu.CDN.OriginTimeout.LoadTimeout <= 0 {
						tmp.Baidu.CDN.OriginTimeout.LoadTimeout = 5
					}
					if tmp.Baidu.CDN.OriginTimeout.ConnectTimeout <= 0 {
						tmp.Baidu.CDN.OriginTimeout.ConnectTimeout = 5
					}

				}

			}

		}

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
