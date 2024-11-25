package baiduyun

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/baidubce/bce-sdk-go/services/cdn"
	"github.com/baidubce/bce-sdk-go/services/cdn/api"
	"github.com/baidubce/bce-sdk-go/services/cert"
	"github.com/projectdiscovery/gologger"
	"github.com/wjlin0/acmeGoBaidu/pkg/config"
	"slices"
	"time"
)

type BaiduYun struct {
	AccessKey  string
	SecretKey  string
	CertClient *cert.Client
	cdnClient  *cdn.Client
}
type SNI struct {
	Enabled bool   `json:"enabled,omitempty"`
	Domain  string `json:"domain,omitempty"`
}

// GetCertListDetail 证书列表详情
func (b *BaiduYun) GetCertListDetail() (*cert.ListCertDetailResult, error) {
	return b.CertClient.ListCertDetail()
}
func (b *BaiduYun) GetCertList() (*cert.ListCertResult, error) {
	return b.CertClient.ListCerts()
}

func (b *BaiduYun) AddCert(privateKey string, certificate string) (*cert.CreateCertResult, error) {
	p, _ := pem.Decode([]byte(certificate))
	if p == nil {
		return nil, errors.New("certificate is not pem")
	}
	parseCertificate, err := x509.ParseCertificate(p.Bytes)
	if err != nil {
		return nil, err
	}
	// 得到证书的域名
	domain := parseCertificate.Subject.CommonName

	certName := domain + "-" + time.Now().Format("2006-01-02")

	return b.CertClient.CreateCert(&cert.CreateCertArgs{
		CertName:        certName,
		CertServerData:  certificate,
		CertPrivateData: privateKey,
	})
}
func (b *BaiduYun) DeleteCert(certId string) error {
	return b.CertClient.DeleteCert(certId)
}

func (b *BaiduYun) CdnList() ([]string, error) {
	domains, _, err := b.cdnClient.ListDomains("")
	return domains, err
}

// CheckDomainInList 判断域名是否在Cdn中
func (b *BaiduYun) CheckDomainInList(domain string) (bool, error) {
	domains, err := b.CdnList()
	if err != nil {
		return false, err
	}
	return slices.Contains(domains, domain), nil
}
func (b *BaiduYun) SetIPv6(domain string, enable bool) error {
	return b.cdnClient.SetIPv6(domain, enable)
}

func (b *BaiduYun) UpdateCdnHTTPSCert(domain string, certId string, http2 bool) error {
	return b.cdnClient.SetDomainHttps(domain, &api.HTTPSConfig{CertId: certId, Enabled: true, Http2Enabled: http2})
}

func (b *BaiduYun) IsValidCdn(domain string) (bool, error) {
	validDomain, err := b.cdnClient.IsValidDomain(domain)
	if err != nil {
		return false, err
	}

	return validDomain.IsValid, nil

}

func (b *BaiduYun) GetCdnHttpsConfig(domain string) (*api.HTTPSConfig, error) {
	return b.cdnClient.GetDomainHttps(domain)
}
func (b *BaiduYun) AddCdn(domain string, baidu *config.Baidu) error {
	_, err := b.cdnClient.CreateDomainWithOptions(domain, baidu.Origin,
		cdn.CreateDomainWithForm(baidu.Form),
		cdn.CreateDomainWithOriginDefaultHost(domain),
		cdn.CreateDomainAsDrcdnType(baidu.Dsa),
	)
	if err != nil {
		return err
	}
	// 设置默认协议跟随 *
	err = b.cdnClient.SetOriginProtocol(domain, "*")
	if err != nil {
		return err
	}
	return nil
}

func (b *BaiduYun) UpdateCdn(domain string, baidu *config.Baidu) error {
	var err error
	// 更新ipv6
	if err = b.cdnClient.SetIPv6(domain, baidu.IPv6); err != nil {
		return fmt.Errorf("set ipv6 error: %v", err)
	}
	// 更新源站
	if len(baidu.Origin) > 0 {
		if err = b.cdnClient.SetDomainOrigin(domain, baidu.Origin, domain); err != nil {
			return fmt.Errorf("set origin error: %v", err)
		}
	}
	// 更新动态加速规则
	if baidu.Dsa != nil && baidu.Dsa.Enabled {
		if err = b.cdnClient.SetDsaConfig(domain, baidu.Dsa); err != nil {
			return fmt.Errorf("set dsa error: %v", err)
		}
		gologger.Info().Msgf("成功设置动态加速规则: %s -> %v", domain, *baidu.Dsa)
	}
	// 更新Seo
	if baidu.Seo != nil {
		if err = b.cdnClient.SetDomainSeo(domain, baidu.Seo); err != nil {
			return fmt.Errorf("set seo error: %v", err)
		}
	}
	if err = b.SetSNI(domain); err != nil {
		return fmt.Errorf("set sni error: %v", err)
	}
	if err = b.cdnClient.SetQUIC(domain, baidu.QUIC); err != nil {
		return fmt.Errorf("set quic error: %v", err)
	}
	if err = b.SetHttp3(domain, baidu.HTTP3); err != nil {
		return fmt.Errorf("set http3 error: %v", err)
	}

	return nil

}
func (b *BaiduYun) SetSNI(domain string) error {
	urlPath := "/v2/domain/" + domain + "/config"
	params := map[string]string{
		"sni": "",
	}
	reqObj := map[string]interface{}{
		"sni": SNI{
			Enabled: true,
			Domain:  domain,
		},
	}

	return b.cdnClient.SendCustomRequest("PUT", urlPath, params, nil, reqObj, nil)
}

func (b *BaiduYun) SetHttp3(domain string, enable bool) error {
	urlPath := "/v2/domain/" + domain + "/config"
	params := map[string]string{
		"http3": "",
	}
	reqObj := map[string]interface{}{
		"http3": map[string]bool{
			"enable": enable,
		},
	}

	return b.cdnClient.SendCustomRequest("PUT", urlPath, params, nil, reqObj, nil)
}
