package aliyun

import (
	"fmt"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"github.com/wjlin0/utils/env"
)

const (
	OssAccessKeyId     = "OSS_ACCESS_KEY_ID"
	OssAccessKeySecret = "OSS_ACCESS_KEY_SECRET"
)

type AliYun struct {
	AccessKey  string
	SecretKey  string
	kodoClient map[string]*oss.Client
}

func NewAliYunFromEnv() (*AliYun, error) {
	var (
		provider credentials.CredentialsProvider = credentials.NewEnvironmentVariableCredentialsProvider()
	)

	m, err := env.Get(OssAccessKeyId, OssAccessKeySecret)
	if err != nil {
		return nil, err
	}
	if m[OssAccessKeyId] == "" || m[OssAccessKeySecret] == "" {
		return nil, fmt.Errorf("accessKey 或 secretKey ")
	}

	kodos := make(map[string]*oss.Client)
	regions := []string{
		"cn-hangzhou",
		"cn-shanghai",
		"cn-nanjing",
		"cn-fuzhou",
		"cn-wuhan-lr",
		"cn-qingdao",
		"cn-beijing",
		"cn-zhangjiakou",
		"cn-huhehaote",
		"cn-wulanchabu",
		"cn-shenzhen",
		"cn-heyuan",
		"cn-guangzhou",
		"cn-chengdu",
		"cn-hongkong",
		"ap-northeast-1",
		"ap-northeast-2",
		"ap-southeast-1",
		"ap-southeast-3",
		"ap-southeast-5",
		"ap-southeast-6",
		"ap-southeast-7",
		"eu-central-1",
		"eu-west-1",
		"us-west-1",
		"us-east-1",
		"me-east-1",
	}

	for _, region := range regions {
		cfg := oss.LoadDefaultConfig().
			WithCredentialsProvider(provider).WithRegion(region)
		client := oss.NewClient(cfg)
		kodos[region] = client
	}

	return &AliYun{
		AccessKey:  m[OssAccessKeyId],
		SecretKey:  m[OssAccessKeySecret],
		kodoClient: kodos,
	}, nil
}
