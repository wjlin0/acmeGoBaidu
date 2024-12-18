package aliyun

import (
	"context"
	"fmt"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/wjlin0/acmeGoBaidu/pkg/certificate"
	"time"
)

func (ali *AliYun) KodoIsBucketExist(bucketName string, region string) (bool, error) {
	// 检查存储空间是否存在
	result, err := ali.kodoClient[region].IsBucketExist(context.TODO(), bucketName)
	if err != nil {
		return false, fmt.Errorf("failed to check if bucket exists %v", err)
	}
	return result, nil
}

func (ali *AliYun) KodoBandCNAME(region string, bucketName, domain string, c certificate.CertificateInfo) error {

	// 创建列出存储空间CNAME的请求
	request := &oss.ListCnameRequest{
		Bucket: oss.Ptr(bucketName),
	}

	// 绑定自定义域名
	result, err := ali.kodoClient[region].ListCname(context.TODO(), request)
	if err != nil {
		return fmt.Errorf("failed to bind cname %v", err)
	}
	// 若已绑定则不再绑定

	for _, cname := range result.Cnames {
		if *cname.Domain == domain {
			// 判断证书是否过期
			if cname.Certificate != nil {
				Date := cname.Certificate.ValidEndDate
				if *Date != "" {
					// 解析时间
					DateTime, err := time.Parse("2006-01-02T15:04:05Z", *Date)
					if err != nil {
						return fmt.Errorf("failed to parse time %v", err)
					}
					// 判断是否过期
					if DateTime.After(time.Now()) {
						return nil
					}
					// 创建添加存储空间CNAME的请求
					requestPut := &oss.PutCnameRequest{
						Bucket: oss.Ptr(bucketName),
						BucketCnameConfiguration: &oss.BucketCnameConfiguration{
							Domain: oss.Ptr(domain),
							CertificateConfiguration: &oss.CertificateConfiguration{
								Force:       oss.Ptr(true),
								Certificate: oss.Ptr(c.Certificate),
								PrivateKey:  oss.Ptr(c.PrivateKey),
							},
						},
					}
					// 绑定自定义域名
					_, err = ali.kodoClient[region].PutCname(context.TODO(), requestPut)
					if err != nil {
						return fmt.Errorf("failed to bind cname %v", err)
					}
					return nil
				}
			}
			// 创建添加存储空间CNAME的请求
			requestPut := &oss.PutCnameRequest{
				Bucket: oss.Ptr(bucketName),
				BucketCnameConfiguration: &oss.BucketCnameConfiguration{
					Domain: oss.Ptr(domain),
					CertificateConfiguration: &oss.CertificateConfiguration{
						Force:       oss.Ptr(true),
						Certificate: oss.Ptr(c.Certificate),
						PrivateKey:  oss.Ptr(c.PrivateKey),
					},
				},
			}
			// 绑定自定义域名
			_, err = ali.kodoClient[region].PutCname(context.TODO(), requestPut)
			if err != nil {
				return fmt.Errorf("failed to bind cname %v", err)
			}
			return nil

		}
	}

	// 创建添加存储空间CNAME的请求
	requestPut := &oss.PutCnameRequest{
		Bucket: oss.Ptr(bucketName),
		BucketCnameConfiguration: &oss.BucketCnameConfiguration{
			Domain: oss.Ptr(domain),
			CertificateConfiguration: &oss.CertificateConfiguration{
				//CertId:      oss.Ptr(fmt.Sprintf("%s-%s", bucketName, domain)),
				Force:       oss.Ptr(true),
				Certificate: oss.Ptr(c.Certificate),
				PrivateKey:  oss.Ptr(c.PrivateKey),
			},
		},
	}
	// 绑定自定义域名
	_, err = ali.kodoClient[region].PutCname(context.TODO(), requestPut)
	if err != nil {
		return fmt.Errorf("failed to bind cname %v", err)
	}
	return nil
}

// KodoCreateCnameToken 创建CnameToken
func (ali *AliYun) KodoCreateCnameToken(region string, bucketName, domain string) (*oss.CnameToken, error) {
	token, _ := ali.KodoGetCnameToken(region, bucketName, domain)
	if token != nil {
		return token, nil
	}
	// 创建Token
	requestToken := &oss.CreateCnameTokenRequest{
		Bucket: oss.Ptr(bucketName),
		BucketCnameConfiguration: &oss.BucketCnameConfiguration{
			Domain: oss.Ptr(domain),
		},
	}

	// 创建Token
	resultToken, err := ali.kodoClient[region].CreateCnameToken(context.TODO(), requestToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create token %v", err)
	}
	return resultToken.CnameToken, nil
}
func (ali *AliYun) KodoGetCnameToken(region string, bucketName, domain string) (*oss.CnameToken, error) {
	// 创建列出存储空间CNAME的请求
	request := &oss.GetCnameTokenRequest{
		Bucket: oss.Ptr(bucketName),
		Cname:  oss.Ptr(domain),
	}
	// 获取Token
	result, err := ali.kodoClient[region].GetCnameToken(context.TODO(), request)
	if err != nil {
		return nil, fmt.Errorf("failed to get token %v", err)
	}
	return result.CnameToken, nil
}
func (ali *AliYun) KodoIsCnameExist(region string, bucketName, domain string) (bool, error) {
	// 创建列出存储空间CNAME的请求
	request := &oss.ListCnameRequest{
		Bucket: oss.Ptr(bucketName),
	}
	// 获取CNAME
	result, err := ali.kodoClient[region].ListCname(context.TODO(), request)
	if err != nil {
		return false, fmt.Errorf("failed to get cname %v", err)
	}
	// 判断是否存在
	for _, cname := range result.Cnames {
		if *cname.Domain == domain {
			return true, nil
		}
	}
	return false, nil
}
