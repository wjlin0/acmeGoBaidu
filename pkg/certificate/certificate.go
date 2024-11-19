package certificate

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"os"
	"time"
)

// CertificateInfo 存储证书信息
type CertificateInfo struct {
	Domain      string    `json:"domain"`
	Certificate string    `json:"certificate"`
	PrivateKey  string    `json:"private_key"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// SaveCertificateInfo 保存证书信息到 JSON 文件
func SaveCertificateInfo(certificates map[string]CertificateInfo, jsonFilePath string) error {
	jsonData, err := json.MarshalIndent(certificates, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化证书信息失败: %v", err)
	}

	err = os.WriteFile(jsonFilePath, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("写入证书信息失败: %v", err)
	}

	return nil
}

// ObtainCertificate 从 ACME 服务器申请证书
func ObtainCertificate(client *lego.Client, domain string) (*certificate.Resource, error) {
	request := certificate.ObtainRequest{
		Domains: []string{domain},
		Bundle:  true,
	}

	certResource, err := client.Certificate.Obtain(request)
	if err != nil {
		return nil, fmt.Errorf("获取证书失败 %s: %v", domain, err)
	}

	return certResource, nil
}

// ParseCertificate 解析 PEM 格式的证书
func ParseCertificate(certData []byte) (*x509.Certificate, error) {
	p, _ := pem.Decode(certData)
	if p == nil {
		return nil, fmt.Errorf("证书内容不是有效的 PEM 格式")
	}

	cert, err := x509.ParseCertificate(p.Bytes)
	if err != nil {
		return nil, fmt.Errorf("无法解析证书: %v", err)
	}

	return cert, nil
}
