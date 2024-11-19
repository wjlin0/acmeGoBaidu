package acme

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"github.com/go-acme/lego/v4/lego"
)

type ACMEClient struct {
	Client *lego.Client
	User   *MyUser
}

// NewACMEClient 创建并返回一个新的 ACME 客户端
func NewACMEClient(email string) (*ACMEClient, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	user := &MyUser{
		Email: email,
		Key:   privateKey,
	}

	legoConfig := lego.NewConfig(user)
	legoConfig.CADirURL = "https://acme-v02.api.letsencrypt.org/directory"
	//legoConfig.CADirURL = "https://acme-staging-v02.api.letsencrypt.org/directory"
	client, err := lego.NewClient(legoConfig)
	if err != nil {
		return nil, err
	}

	return &ACMEClient{
		Client: client,
		User:   user,
	}, nil
}
