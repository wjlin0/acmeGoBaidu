package baiduyun

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/baidubce/bce-sdk-go/services/cdn"
	"github.com/baidubce/bce-sdk-go/services/cert"
	"github.com/go-acme/lego/v4/platform/config/env"
)

const (
	BAIDUYUN_ACCESSKEY = "BAIDUYUN_ACCESSKEY"
	BAIDUYUN_SECRETKEY = "BAIDUYUN_SECRETKEY"
	BAIDUYUN_REGION    = "BAIDUYUN_REGION"
)

func NewBaiduYun(accessKey, secretKey string) (*BaiduYun, error) {
	if accessKey == "" || secretKey == "" {
		return NewBaiduYunFromEnv()
	}
	return &BaiduYun{
		AccessKey: accessKey,
		SecretKey: secretKey,
	}, nil
}

func NewBaiduYunFromEnv() (b *BaiduYun, err error) {
	m, err := env.Get(BAIDUYUN_ACCESSKEY, BAIDUYUN_SECRETKEY)
	if err != nil {
		return nil, err
	}
	if m[BAIDUYUN_ACCESSKEY] == "" || m[BAIDUYUN_SECRETKEY] == "" {
		return nil, fmt.Errorf("accessKey 或 secretKey ")
	}
	certClient, err := cert.NewClient(m[BAIDUYUN_ACCESSKEY], m[BAIDUYUN_SECRETKEY], "https://certificate.baidubce.com")
	if err != nil {
		return nil, err
	}
	cdnClient, err := cdn.NewClient(m[BAIDUYUN_ACCESSKEY], m[BAIDUYUN_SECRETKEY], "https://cdn.baidubce.com")
	return &BaiduYun{
		AccessKey:  m[BAIDUYUN_ACCESSKEY],
		SecretKey:  m[BAIDUYUN_SECRETKEY],
		CertClient: certClient,
		cdnClient:  cdnClient,
	}, nil

}

func hmacSHA256Hex(key string, message string) string {
	// 调用HMAC SHA256算法，根据开发者提供的密钥（key）和密文（message）输出密文摘要，并把结果转换为小写形式的十六进制字符串。
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}
