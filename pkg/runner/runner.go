package runner

import (
	"encoding/json"
	"fmt"
	"github.com/baidubce/bce-sdk-go/services/cert"
	"github.com/go-acme/lego/v4/providers/dns"
	"github.com/go-acme/lego/v4/registration"
	"github.com/projectdiscovery/gologger"
	baidudns "github.com/wjlin0/acmeGoBaidu/pkg/baidu"
	"github.com/wjlin0/acmeGoBaidu/pkg/baiduyun"
	"github.com/wjlin0/acmeGoBaidu/pkg/types"
	"os"
	"time"

	"github.com/wjlin0/acmeGoBaidu/pkg/acme"
	"github.com/wjlin0/acmeGoBaidu/pkg/certificate"
	"github.com/wjlin0/acmeGoBaidu/pkg/config"
)

// Runner 结构体保存所有证书申请的相关信息
type Runner struct {
	Options      *types.Options
	Config       config.Config
	Client       *acme.ACMEClient
	Certificates map[string]certificate.CertificateInfo
	JsonFilePath string
	Baidu        *baiduyun.BaiduYun
}

// NewRunner 创建一个新的 Runner 实例
func NewRunner(opts *types.Options) (*Runner, error) {
	// 加载配置文件
	c, err := config.LoadConfig(opts.ConfigFile)
	if err != nil {
		return nil, fmt.Errorf("无法加载配置: %v", err)
	}

	// 创建 ACME 客户端
	AcmeClient, err := acme.NewACMEClient(c.Acme.Email)
	if err != nil {
		return nil, fmt.Errorf("创建 ACME 客户端失败: %v", err)
	}

	// 加载现有证书信息
	certificates := make(map[string]certificate.CertificateInfo)
	if _, err := os.Stat(opts.JsonPath); err == nil {
		fileData, err := os.ReadFile(opts.JsonPath)
		if err == nil {
			err := json.Unmarshal(fileData, &certificates)
			if err != nil {
				return nil, fmt.Errorf("加载证书信息失败: %v", err)
			}
		}
	}

	baidu, err := baiduyun.NewBaiduYunFromEnv()
	if err != nil {
		return nil, fmt.Errorf("创建百度云客户端失败: %v", err)
	}

	return &Runner{
		Config:       c,
		Client:       AcmeClient,
		Certificates: certificates,
		JsonFilePath: opts.JsonPath,
		Baidu:        baidu,
		Options:      opts,
	}, nil
}

// Run 执行证书申请流程
func (r *Runner) Run() error {
	// 遍历配置中的域名，申请证书
	for _, domainConfig := range r.Config.Domains {
		domain := domainConfig.Domain
		providerName := domainConfig.Provider

		// 检查现有证书是否有效
		if c, exists := r.Certificates[domain]; exists {
			if time.Until(c.ExpiresAt).Hours() > 14*24 {
				gologger.Warning().Msgf("证书已存在且有效，跳过域名: %s\n", domain)
				continue
			}
		}

		// 创建 DNS 提供商挑战
		provider, err := dns.NewDNSChallengeProviderByName(providerName)
		if err != nil {
			gologger.Error().Msgf("无法创建 DNS 提供商 %s 的挑战: %v", providerName, err)
			continue
		}

		err = r.Client.Client.Challenge.SetDNS01Provider(provider)
		if err != nil {
			gologger.Error().Msgf("设置 DNS 提供商 %s 失败: %v", providerName, err)
			continue
		}

		// 注册并申请证书
		reg, err := r.Client.Client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
		if err != nil {
			gologger.Error().Msgf("注册失败: %v", err)
		}
		r.Client.User.Registration = reg

		certResource, err := certificate.ObtainCertificate(r.Client.Client, domain)
		if err != nil {
			gologger.Error().Msgf("申请证书失败: %v", err)
			continue
		}

		// 解析证书
		c, err := certificate.ParseCertificate([]byte(certResource.Certificate))
		if err != nil {
			gologger.Error().Msgf("解析证书失败: %v", err)
			continue
		}

		// 保存证书信息
		r.Certificates[domain] = certificate.CertificateInfo{
			Domain:      domain,
			Certificate: string(certResource.Certificate),
			PrivateKey:  string(certResource.PrivateKey),
			ExpiresAt:   c.NotAfter,
		}

		gologger.Info().Msgf("成功申请证书: %s", domain)

	}
	err := r.Output()
	if err != nil {
		return err
	}
	err = r.UpdateBaiduCdnCertificate()
	if err != nil {
		return err
	}

	gologger.Info().Msg("证书申请完成")
	return nil
}

func (r *Runner) UpdateBaiduCdnCertificate() error {
	details, err := r.Baidu.GetCertListDetail()
	if err != nil {
		return err
	}
	for _, domainConfig := range r.Config.Domains {
		domain := domainConfig.Domain
		provider := domainConfig.Provider
		cname := domainConfig.Baidu.Cname
		// 如果没有配置百度云CDN，跳过
		if domainConfig.Baidu == nil {
			continue
		}
		if c, exists := r.Certificates[domain]; exists {
			// 首先判断CDN配置是否存在
			ok, err := r.Baidu.IsValidCdn(domain)
			if err != nil {

				continue
			}
			if ok {
				// 添加CDN配置
				err := r.Baidu.AddCdn(domain, domainConfig.Baidu)
				if err != nil {
					gologger.Error().Msgf("添加CDN配置失败: %v", err)
					continue
				}

				gologger.Info().Msgf("成功添加CDN配置: %s", domain)

			}

			// 判断证书是否存在
			f := false
			var certMeta cert.CertificateDetailMeta
			for _, v := range details.Certs {
				if v.CertCommonName == domain {
					f = true
					certMeta = v
					break
				}
			}
			var certId string

			if !f {
				// 证书不存在，创建证书
				gologger.Info().Msgf("证书不存在，创建证书: %s", domain)
				certResult, err := r.Baidu.AddCert(c.PrivateKey, c.Certificate)
				if err != nil {
					continue
				}
				certId = certResult.CertId
				gologger.Info().Msgf("成功创建证书: %s", domain)
				// 更新CDN证书
				err = r.Baidu.UpdateCdnHTTPSCert(domain, certId)
				if err != nil {
					gologger.Info().Msgf("更新CDNHTTPS配置失败: %v", err)
					continue
				}

			} else {
				// 证书存在，更新证书
				gologger.Info().Msgf("证书存在，更新证书: %s", domain)
				parse, _ := time.Parse(time.RFC3339, certMeta.CertStopTime)
				if parse.Unix() < c.ExpiresAt.Unix() || (parse.Unix() > time.Now().Unix()) == false {
					gologger.Info().Msgf("证书已过期，更新证书: %s", domain)
					certResult, err := r.Baidu.AddCert(c.PrivateKey, c.Certificate)
					if err != nil {
						gologger.Error().Msgf("更新证书失败: %v", err)
						continue
					}
					gologger.Info().Msgf("成功更新证书: %s", domain)
					err = r.Baidu.UpdateCdnHTTPSCert(domain, certResult.CertId)
					if err != nil {
						gologger.Info().Msgf("更新CDNHTTPS配置失败: %v", err)
						continue
					}
					// 删除旧证书
					gologger.Info().Msgf("删除旧证书: %s -> %s", domain, certMeta.CertId)
					err = r.Baidu.DeleteCert(certMeta.CertId)
					if err != nil {
						gologger.Error().Msgf("删除旧证书失败: %v", err)
					}
				} else {
					gologger.Info().Msgf("证书未过期，跳过更新: %s", domain)
					err = r.Baidu.UpdateCdnHTTPSCert(domain, certMeta.CertId)
					if err != nil {
						gologger.Error().Msgf("更新CDN证书失败: %v", err)
						continue
					}
				}
			}
			var providerDNS baidudns.Provider

			// 配置域名的CNAME解析
			if cname {
				if providerDNS, err = baidudns.NewDNSChallengeProviderByName(provider); err != nil {
					gologger.Error().Msgf("无法创建 DNS 提供商 %s 的挑战: %v", provider, err)
					continue
				}
				if err = providerDNS.CreateCNAMERecord(domain); err != nil {
					gologger.Error().Msgf("创建CNAME记录失败: %v", err)
					continue
				}
			}
		}
	}
	return nil
}
