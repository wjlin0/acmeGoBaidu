package baiduyun

import "testing"

func TestBaiduYun_GetSSLList(t *testing.T) {
	baidu, err := NewBaiduYunFromEnv()
	if err != nil {
		t.Fatal(err)
		return
	}
	list, err := baidu.GetCertListDetail()
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Log(list)

}
