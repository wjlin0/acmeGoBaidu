package acme

import (
	"crypto"
	"github.com/go-acme/lego/v4/registration"
)

// MyUser 定义一个用户结构，用于 ACME 注册
type MyUser struct {
	Email        string
	Registration *registration.Resource
	Key          crypto.PrivateKey
}

func (u *MyUser) GetEmail() string {
	return u.Email
}

func (u *MyUser) GetRegistration() *registration.Resource {
	return u.Registration
}

func (u *MyUser) GetPrivateKey() crypto.PrivateKey {
	return u.Key
}
