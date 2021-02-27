package main

import (
	"github.com/go-ini/ini"
)

var (
	CONFIG_FILE = "config.ini"
)

var Config LogConfig

type LogConfig struct {
	Mode        string // 开发模式
	ExePath     string // exe的路径
	ZipPath     string // 备份的压缩文件存储路径
	LocalIpPath string // 外网ip地址
	SqlBinPath  string // mysql的bin目录
	Mysql       MysqlConfig
	Times       TimeConfig
}

type MysqlConfig struct {
	DBType        string
	DBUser        string
	DBPasswd      string
	DBHost1       string
	DBHost2       string
	DBNameA       string
	DBNameB       string
	DBTablePrefix string
}

type TimeConfig struct {
	Hour   string
	Minute string
	Second string
}

func ReadConfig() error {
	cfg, err := ini.Load(CONFIG_FILE)
	if err != nil {
		return err
	}

	Config.Mode = cfg.Section("").Key("app_mode").String()
	// 因为服务的安装必须要管理员权限，安装完成之后
	Config.ExePath = "" //cfg.Section("paths").Key("exe_path").String()
	Config.ZipPath = cfg.Section("paths").Key("zip_disk").String()
	Config.SqlBinPath = cfg.Section("paths").Key("sql_bin_path").String()
	Config.LocalIpPath = cfg.Section("paths").Key("local_ip").String()

	err = ReadMysqlConfig(cfg.Section("database"), &Config.Mysql)
	if err != nil {
		return err
	}

	err = ReadTimeConfig(cfg.Section("times"), &Config.Times)
	if err != nil {
		return err
	}

	return nil
}

func ReadMysqlConfig(section *ini.Section, mc *MysqlConfig) error {
	mc.DBType = section.Key("DBType").String()
	mc.DBUser = section.Key("DBUser").String()
	mc.DBPasswd = section.Key("DBPasswd").String()
	mc.DBHost1 = section.Key("DBHost1").String()
	mc.DBHost2 = section.Key("DBHost2").String()
	mc.DBNameA = section.Key("DBNameA").String()
	mc.DBNameB = section.Key("DBNameB").String()
	mc.DBTablePrefix = "" //section.Key("DBTablePrefix").String()

	return nil
}

func ReadTimeConfig(section *ini.Section, mc *TimeConfig) error {
	mc.Hour = section.Key("Hour").String()
	mc.Minute = section.Key("Minute").String()
	mc.Second = section.Key("Second").String()

	return nil
}
