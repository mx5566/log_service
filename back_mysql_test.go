package main

import "testing"

func TestBackupMySqlDb1(t *testing.T) {
	err, path := BackupMySqlDb1("127.0.0.1", "3306", "root", "123456", "log", "log_20210204", "")

	if err != nil {
		t.Logf("err[%s] path[%s]", err.Error(), path)
	} else {
		t.Logf("err[%s] path[%s]", "", path)
	}

	err = Zip(path, path+".zip")
	if err != nil {
		t.Logf("zip err[%s]", err.Error())
	} else {
		t.Logf("zip no error")
	}
}
