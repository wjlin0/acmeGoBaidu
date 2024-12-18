package runner

import (
	"encoding/json"
	"fmt"
	"github.com/baidubce/bce-sdk-go/services/cert"
	"github.com/go-acme/lego/v4/providers/dns"
	"github.com/go-acme/lego/v4/registration"
	"github.com/projectdiscovery/gologger"
	"github.com/wjlin0/acmeGoBaidu/pkg/acme"
	baidudns "github.com/wjlin0/acmeGoBaidu/pkg/baidu"
	"github.com/wjlin0/acmeGoBaidu/pkg/baidu/dns01"
	"github.com/wjlin0/acmeGoBaidu/pkg/certificate"
	"github.com/wjlin0/acmeGoBaidu/pkg/config"
	"github.com/wjlin0/acmeGoBaidu/pkg/types"
	"github.com/wjlin0/acmeGoBaidu/pkg/yun/aliyun"
	"github.com/wjlin0/acmeGoBaidu/pkg/yun/baiduyun"
	"os"
	"strings"
	"time"
)

// Runner 结构体保存所有证书申请的相关信息
type Runner struct {
	Options      *types.Options
	Config       config.Config
	Client       *acme.ACMEClient
	Certificates map[string]certificate.CertificateInfo
	JsonFilePath string
	Baidu        *baiduyun.BaiduYun
	AliYun       *aliyun.AliYun
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

	ali, err := aliyun.NewAliYunFromEnv()
	if err != nil {
		return nil, fmt.Errorf("创建阿里云客户端失败: %v", err)
	}
	return &Runner{
		Config:       c,
		Client:       AcmeClient,
		Certificates: certificates,
		JsonFilePath: opts.JsonPath,
		Baidu:        baidu,
		Options:      opts,
		AliYun:       ali,
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
	err = r.UpdateAliYun()
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
		// 如果没有配置百度云CDN，跳过
		if strings.HasPrefix(domainConfig.To, "baidu") == false {
			continue
		}
		if domainConfig.Baidu == nil {
			continue
		}
		if domainConfig.Baidu.CDN == nil {
			continue
		}
		cnameInfo := domainConfig.Baidu.CDN.CnameInfo
		if c, exists := r.Certificates[domain]; exists {
			// 首先判断CDN配置是否存在
			ok, err := r.Baidu.IsValidCdn(domain)
			if err != nil {
				continue
			}
			if ok {
				// 添加CDN配置
				err := r.Baidu.AddCdn(domain, domainConfig.Baidu.CDN)
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
				err = r.Baidu.UpdateCdnHTTPSCert(domain, certId, domainConfig.Baidu.CDN.HTTP2)
				if err != nil {
					gologger.Info().Msgf("更新CDNHTTPS配置失败: %v", err)
					continue
				}

			} else {
				// 证书存在，更新证书

				parse, _ := time.Parse(time.RFC3339, certMeta.CertStopTime)
				if parse.Unix() < c.ExpiresAt.Unix() || (parse.Unix() > time.Now().Unix()) == false {
					gologger.Info().Msgf("证书已过期，更新证书: %s", domain)
					certResult, err := r.Baidu.AddCert(c.PrivateKey, c.Certificate)
					if err != nil {
						gologger.Error().Msgf("更新证书失败: %v", err)
						continue
					}
					gologger.Info().Msgf("成功更新证书: %s", domain)
					err = r.Baidu.UpdateCdnHTTPSCert(domain, certResult.CertId, domainConfig.Baidu.CDN.HTTP2)
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
					err = r.Baidu.UpdateCdnHTTPSCert(domain, certMeta.CertId, domainConfig.Baidu.CDN.HTTP2)
					if err != nil {
						gologger.Error().Msgf("更新CDN证书失败: %v", err)
						continue
					}
				}
			}
			if err = r.Baidu.UpdateCdn(domain, domainConfig.Baidu.CDN); err != nil {
				gologger.Error().Msgf("更新CDN配置失败: %v", err)
			}
			var providerDNS baidudns.Provider

			// 配置域名的CNAME解析
			if cnameInfo != nil && cnameInfo.Enabled {
				if providerDNS, err = baidudns.NewDNSChallengeProviderByName(provider); err != nil {
					gologger.Error().Msgf("无法创建 DNS 提供商 %s 的挑战: %v", provider, err)
					continue
				}

				//dnsType, value, err := dns01.CheckCNAMExistBaidu(dns01.ToFqdn(domain))
				ok, value, err := providerDNS.ExistsRecord("CNAME", domain)
				if err != nil {
					gologger.Error().Msgf("检查CNAME记录失败: %v", err)
					continue
				}
				if !ok {
					ok, value, err = providerDNS.ExistsRecord("A", domain)
					if err != nil {
						gologger.Error().Msgf("检查A记录失败: %v", err)
						continue
					}
					if ok {
						gologger.Info().Msgf("A记录已存在, 正在删除: %s -> %s", domain, value)
						// 删除旧A记录
						if err = providerDNS.DeleteRecord("A", domain); err != nil {
							gologger.Error().Msgf("删除A记录失败: %v", err)
							continue
						}
					}
					ok, value, err = providerDNS.ExistsRecord("AAAA", domain)
					if err != nil {
						gologger.Error().Msgf("检查AAAA记录失败: %v", err)
						continue
					}
					if ok {
						gologger.Info().Msgf("AAAA记录已存在, 正在删除: %s -> %s", domain, value)
						// 删除旧A记录
						if err = providerDNS.DeleteRecord("AAAA", domain); err != nil {
							gologger.Error().Msgf("删除AAAA记录失败: %v", err)
							continue
						}
					}

					if err = providerDNS.CreateRecord("CNAME", domain, dns01.ToFqdn(cnameInfo.Value)); err != nil {
						gologger.Error().Msgf("创建CNAME记录失败: %v", err)
						continue
					}
					gologger.Info().Msgf("成功创建CNAME记录: %s -> %s.a.bdydns.com", domain, domain)
					continue
				} else {
					if dns01.UnFqdn(cnameInfo.Value) == dns01.UnFqdn(cnameInfo.Value) {
						gologger.Info().Msgf("CNAME记录已存在: %s -> %s", domain, value)
						continue
					}
					gologger.Info().Msgf("CNAME记录已存在, 正在更新: %s -> %s -> %s", domain, value, cnameInfo.Value)
					// 删除旧CNAME记录
					if err = providerDNS.DeleteRecord("CNAME", domain); err != nil {
						gologger.Error().Msgf("删除CNAME记录失败: %v", err)
						continue
					}
					if err = providerDNS.CreateRecord("CNAME", domain, dns01.ToFqdn(cnameInfo.Value)); err != nil {
						gologger.Error().Msgf("创建CNAME记录失败: %v", err)
						continue
					}
					gologger.Info().Msgf("成功更新CNAME记录: %s -> %s", domain, cnameInfo.Value)
				}
			}

		}
	}
	return nil
}

func (r *Runner) UpdateAliYun() error {
	for _, domainConfig := range r.Config.Domains {
		provider := domainConfig.Provider
		domain := domainConfig.Domain
		if strings.HasPrefix(domainConfig.To, "ali") == false {
			continue
		}
		// action
		action := strings.Split(domainConfig.To, ",")
		if len(action) != 2 {
			action = []string{"cdn", "kodo"}
		}
		if domainConfig.AliYun == nil {
			continue
		}
		if c, exists := r.Certificates[domainConfig.Domain]; exists {
			switch action[1] {
			case "kodo":

				// 首先判断存储桶是否存在
				ok, err := r.AliYun.KodoIsBucketExist(domainConfig.AliYun.Kodo.Bucket, domainConfig.AliYun.Kodo.Region)
				if err != nil {
					gologger.Error().Msgf("检查存储桶是否存在失败: %v", err)
					continue
				}
				if !ok {
					continue
				}
				if ok, _ = r.AliYun.KodoIsCnameExist(domainConfig.AliYun.Kodo.Region, domainConfig.AliYun.Kodo.Bucket, domain); !ok {
					token, err := r.AliYun.KodoCreateCnameToken(domainConfig.AliYun.Kodo.Region, domainConfig.AliYun.Kodo.Bucket, domain)
					if err != nil {
						gologger.Error().Msgf("创建CnameToken失败: %v", err)
						continue
					}

					err = createTxt(provider, fmt.Sprintf("_dnsauth.%s", domain), *token.Token)
					if err != nil {
						return err
					}
					defer func() {
						_ = deleteTxt(provider, fmt.Sprintf("_dnsauth.%s", domain), *token.Token)
					}()

					time.Sleep(10 * time.Second)
				}

				err = r.AliYun.KodoBandCNAME(domainConfig.AliYun.Kodo.Region, domainConfig.AliYun.Kodo.Bucket, domainConfig.Domain, c)
				if err != nil {

					gologger.Error().Msgf("绑定CNAME失败: %v", err)
					continue
				}

				gologger.Info().Msgf("成功绑定CNAME: %s", domainConfig.Domain)
				var providerDNS baidudns.Provider
				cnameInfo := domainConfig.AliYun.Kodo.CnameInfo
				// 配置域名的CNAME解析
				if cnameInfo != nil && cnameInfo.Enabled {
					if cnameInfo.Value == "" {
						cnameInfo.Value = fmt.Sprintf("%s.oss-%s.aliyuncs.com.", domainConfig.AliYun.Kodo.Bucket, domainConfig.AliYun.Kodo.Region)
					}

					if providerDNS, err = baidudns.NewDNSChallengeProviderByName(provider); err != nil {
						gologger.Error().Msgf("无法创建 DNS 提供商 %s 的挑战: %v", provider, err)
						continue
					}

					//dnsType, value, err := dns01.CheckCNAMExistBaidu(dns01.ToFqdn(domain))
					ok, value, err := providerDNS.ExistsRecord("CNAME", domain)
					if err != nil {
						gologger.Error().Msgf("检查CNAME记录失败: %v", err)
						continue
					}
					if !ok {
						ok, value, err = providerDNS.ExistsRecord("A", domain)
						if err != nil {
							gologger.Error().Msgf("检查A记录失败: %v", err)
							continue
						}
						if ok {
							gologger.Info().Msgf("A记录已存在, 正在删除: %s -> %s", domain, value)
							// 删除旧A记录
							if err = providerDNS.DeleteRecord("A", domain); err != nil {
								gologger.Error().Msgf("删除A记录失败: %v", err)
								continue
							}
						}
						ok, value, err = providerDNS.ExistsRecord("AAAA", domain)
						if err != nil {
							gologger.Error().Msgf("检查AAAA记录失败: %v", err)
							continue
						}
						if ok {
							gologger.Info().Msgf("AAAA记录已存在, 正在删除: %s -> %s", domain, value)
							// 删除旧A记录
							if err = providerDNS.DeleteRecord("AAAA", domain); err != nil {
								gologger.Error().Msgf("删除AAAA记录失败: %v", err)
								continue
							}
						}

						if err = providerDNS.CreateRecord("CNAME", domain, dns01.ToFqdn(cnameInfo.Value)); err != nil {
							gologger.Error().Msgf("创建CNAME记录失败: %v", err)
							continue
						}
						gologger.Info().Msgf("成功创建CNAME记录: %s -> %s.a.bdydns.com", domain, domain)
						continue
					} else {
						if dns01.UnFqdn(cnameInfo.Value) == dns01.UnFqdn(cnameInfo.Value) {
							gologger.Info().Msgf("CNAME记录已存在: %s -> %s", domain, value)
							continue
						}
						gologger.Info().Msgf("CNAME记录已存在, 正在更新: %s -> %s -> %s", domain, value, cnameInfo.Value)
						// 删除旧CNAME记录
						if err = providerDNS.DeleteRecord("CNAME", domain); err != nil {
							gologger.Error().Msgf("删除CNAME记录失败: %v", err)
							continue
						}
						if err = providerDNS.CreateRecord("CNAME", domain, dns01.ToFqdn(cnameInfo.Value)); err != nil {
							gologger.Error().Msgf("创建CNAME记录失败: %v", err)
							continue
						}
						gologger.Info().Msgf("成功更新CNAME记录: %s -> %s", domain, cnameInfo.Value)
					}
				}

			}
		}
	}

	return nil
}

func createTxt(provider string, domain, txt string) error {
	providerDNS, err := baidudns.NewDNSChallengeProviderByName(provider)
	if err != nil {
		return fmt.Errorf("无法创建 DNS 提供商 %s 的挑战: %v", provider, err)
	}

	//dnsType, value, err := dns01.CheckCNAMExistBaidu(dns01.ToFqdn(domain))
	err = providerDNS.CreateRecord("TXT", domain, txt)
	if err != nil {
		return fmt.Errorf("创建TXT记录失败: %v", err)
	}
	return nil
}
func deleteTxt(provider string, domain, txt string) error {
	providerDNS, err := baidudns.NewDNSChallengeProviderByName(provider)
	if err != nil {
		return fmt.Errorf("无法创建 DNS 提供商 %s 的挑战: %v", provider, err)
	}

	//dnsType, value, err := dns01.CheckCNAMExistBaidu(dns01.ToFqdn(domain))
	err = providerDNS.DeleteRecord("TXT", domain)
	if err != nil {
		return fmt.Errorf("删除TXT记录失败: %v", err)
	}
	return nil
}
