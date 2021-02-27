package main

import "testing"

func TestReadConfig(t *testing.T) {
	err := ReadConfig()
	if err != nil {
		t.Errorf("%s", err.Error())
		return
	}

	t.Log(Config)
}
