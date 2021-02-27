package main

import (
	"fmt"
	"github.com/mx5566/logm"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
)

/**
 *
 * 备份MySql数据库
 * @param 	host: 			数据库地址: localhost
 * @param 	port:			端口: 3306
 * @param 	user:			用户名: root
 * @param 	password:		密码: root
 * @param 	databaseName:	需要被分的数据库名: test
 * @param 	tableName:		需要备份的表名: user
 * @param 	sqlPath:		备份SQL存储路径: D:/backup/test/
 * @return 	backupPath
 *
 */
func BackupMySqlDb(host, port, user, password, databaseName, tableName, sqlPath string) (error, string) {
	var cmd *exec.Cmd

	wd, _ := os.Getwd()
	name := wd + BackFileName("")
	fmt.Println(name)

	binPath := "C:\\Program Files\\MySQL\\MySQL Server 8.0\\bin"

	if tableName == "" {
		cmd = exec.Command(binPath+"\\mysqldump.exe", "--opt", "-h"+host, "-P"+port, "-u"+user, "-p"+password, databaseName, ">", name)
	} else {
		cmd = exec.Command(binPath+"\\mysqldump.exe", "--opt", "-h"+host, "-P"+port, "-u"+user, "-p"+password, databaseName, tableName, ">", name)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
		return err, ""
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
		return err, ""
	}

	bytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		log.Fatal(err)
		return err, ""
	}
	//now := time.Now().Format("20060102150405")
	ip, err := ExternalIP()
	if err != nil {
		logm.ErrorfE("get ip error [%s]", err.Error())
		ip = net.IPv4(byte(127), byte(0), byte(0), byte(1))
	}

	var backupPath string
	if tableName == "" {
		backupPath = sqlPath + ip.String() + ".sql"
	} else {
		backupPath = sqlPath + ip.String() + ".sql"
	}
	err = ioutil.WriteFile(backupPath, bytes, 0644)

	if err != nil {
		return err, ""
	}
	return nil, backupPath
}

func BackupMySqlDb1(host, port, user, password, databaseName, tableName, sqlPath string) (error, string) {
	fileName := BackFileName(tableName)
	wd, _ := os.Getwd()
	fullName := wd + "\\" + fileName

	var err error
	if tableName == "" {
		err = ExecCmd("/C", Config.SqlBinPath+"\\mysqldump.exe", "--opt", "-h"+host, "-P"+port, "-u"+user, "-p"+password, databaseName, ">", fullName)
	} else {
		err = ExecCmd("/C", Config.SqlBinPath+"\\mysqldump.exe", "--opt", "-h"+host, "-P"+port, "-u"+user, "-p"+password, databaseName, tableName, ">", fullName)
	}

	if err != nil {
		return err, ""
	}

	return err, fileName
}

func BackFileName(tblName string) string {
	ip, err := GetIp()
	if err != nil {
		logm.ErrorfE("get ip error [%s]", err.Error())
		ip = net.IPv4(byte(127), byte(0), byte(0), byte(1))
	}

	backupPath := ip.String() + "_" + tblName + ".sql"

	return backupPath
}
