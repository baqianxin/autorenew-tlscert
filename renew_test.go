package main

import "testing"

func Test_PublishCert_Test(t *testing.T) {
	// https://dash.home.qianxin.me:9180
	err := PublishCert("/Users/qianxinqba/.acme.sh/*.sec.qianxin.me_ecc", "http://dash.home.qianxin.me:9180", "edd1c9f034335f136f87ad84b625c8f1", "*.sec.qianxin.me", true, true)
	t.Log("Test_PublishCert_Test", err)
}
