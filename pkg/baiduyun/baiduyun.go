package baiduyun

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"github.com/baidubce/bce-sdk-go/services/cdn"
	"github.com/baidubce/bce-sdk-go/services/cdn/api"
	"github.com/baidubce/bce-sdk-go/services/cert"
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

func (b *BaiduYun) UpdateCdnHTTPSCert(domain string, certId string) error {
	if err := b.cdnClient.SetDomainHttps(domain, &api.HTTPSConfig{CertId: certId, Enabled: true, Http2Enabled: true}); err != nil {
		return err
	}

	return b.cdnClient.SetQUIC(domain, true)
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
