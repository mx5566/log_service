package main

import (
	"testing"
)

func TestExternalIP(t *testing.T) {
	ip, err := ExternalIP()
	if err != nil {
		t.Log(err)
	}

	t.Log(ip.String())
}

func TestReadLocalIp(t *testing.T) {
	_ = ReadConfig()
	ip, err := GetIp()
	if err != nil {
		t.Log(err)
		return
	}

	t.Log(ip.String())
}

func TestSplitLogTableName(t *testing.T) {
	year, month, day, err := SplitLogTableName("log_20120212")

	t.Logf("year[%s] month[%s] day[%s] err[%v]", year, month, day, err)
}
