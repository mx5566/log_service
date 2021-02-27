package main

import (
	"flag"
	"fmt"
	"github.com/kardianos/service"
	"github.com/mx5566/logm"
	"github.com/robfig/cron/v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"os"
	"runtime"
	"strconv"
	"time"
)

var serviceConfig = &service.Config{
	Name:        "log_bak_service",
	DisplayName: "log_service",
	Description: "bak log database",
}

type FlagParams struct {
	Dir      string // 路径
	Operator string // 操作类型 install start1 stop uninstall restart
	Daemon   bool   // 是否后台执行
}

var Params FlagParams

var GDB1 *gorm.DB
var GDB2 *gorm.DB

func init() {
	InitFlag()
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			switch err.(type) {
			case runtime.Error: // 运行时错误
				logm.TraceErrorEx("log_service recover runtime err[%v]", err)
			default: // 非运行时错误
				logm.TraceErrorEx("log_service recover not runtime err[%v]", err)
			}
		}
	}()

	// 解析命令行参数
	flag.Parse()

	// log_service.exe -d true -dir C:\log_service\ -cmd install
	// log_service.exe -d true  -cmd start1
	if Params.Daemon {
		if Params.Operator == "" && Params.Dir == "" {
			Params.Daemon = false
		}
	}

	if Params.Operator == "install" {
		if Params.Dir == "" {
			os.Exit(0)
		}

		serviceConfig.Arguments = append(serviceConfig.Arguments, "-dir")
		serviceConfig.Arguments = append(serviceConfig.Arguments, Params.Dir)
	}

	// 切换目录因为如果服务安装之后(管理员权限) 目录就会在system32目录下面
	if Params.Dir != "" {
		err := os.Chdir(Params.Dir)
		if err != nil {
			fmt.Printf("切换目录失败[%s]", err.Error())
			os.Exit(0)
		}
	}

	cur, _ := os.Getwd()
	fmt.Println("当前的目录->", cur)

	// 日志模块初始化
	logm.Init("log_service", map[string]string{"errFile": "log_service_err.log", "logFile": "log_service.log"}, "debug")
	logm.DebugfE("日志模块初始化结束")

	// 构建服务对象
	prog := &Program{}

	s, err := service.New(prog, serviceConfig)
	if err != nil {
		log.Fatal(err)
		return
	}

	logm.DebugfE("", os.Args)
	if !Params.Daemon || Params.Operator == "" {
		err = s.Run()
		if err != nil {
			logm.ErrorfE(err.Error())
		}
		return
	}

	cmd := Params.Operator

	fmt.Println("服务命令", cmd)

	if cmd == "install" {
		err = s.Install()
		if err != nil {
			// 文本日志
			logm.ErrorfE("安装服务出错[%s]", err.Error())
			return
		}
		logm.DebugfE("安装成功")
		return
	}

	if cmd == "uninstall" {
		status, err := s.Status()
		if status == service.StatusRunning {
			err = s.Stop()
		}

		err = s.Uninstall()
		if err != nil {
			// 文本日志
			logm.ErrorfE("卸载服务出错[%s]", err.Error())
			return
		}
		logm.DebugfE("卸载成功")
		return
	}

	if cmd == "start1" {
		err = s.Start()
		if err != nil {
			// 文本日志
			logm.ErrorfE("启动服务出错[%s]", err.Error())
			return
		}

		logm.DebugfE("启动成功")
		return
	}

	if cmd == "stop" {
		err = s.Stop()
		if err != nil {
			// 文本日志
			logm.ErrorfE("停止服务出错[%s]", err.Error())
			return
		}

		logm.DebugfE("停止成功")
		return
	}

	if cmd == "restart" {
		status, err := s.Status()
		if status == service.StatusRunning {
			err = s.Restart()
		} else {
			err = s.Start()
		}

		if err != nil {
			// 文本日志
			logm.ErrorfE("重启服务出错[%s]", err.Error())
			return
		}
		logm.DebugfE("重启成功")
		return
	}

	if cmd == "status" {
		status, err := s.Status()
		if err != nil {
			// 文本日志
			logm.ErrorfE("查看服务状态出错[%s]", err.Error())
			return
		}

		logm.DebugfE("服务状态[%d]", status)
		return
	}
}

type Program struct{}

func (p *Program) Start(s service.Service) error {
	logm.DebugfE("开始服务")
	err := Init()
	if err != nil {
		return err
	}

	go p.run()
	return nil
}

func (p *Program) Stop(s service.Service) error {
	logm.DebugfE("停止服务")

	// 触发一个信号关闭定时任务

	return nil
}

func (p *Program) run() {
	// 每天的00:00:01检测本机的log数据库里面的表，有没有满足条件的需要先备份在删除 原来的删除表的存储过程不在通过事件来调用
	logm.DebugfE("运行服务run")

	//spec := "1 0 0 * * ?" //每天早上00:00:01执行一次
	spec := Config.Times.Second + " " + Config.Times.Minute + " " + Config.Times.Hour + " * * ?"
	//spec := "0 */2 * * * ?" //每分钟执行一次

	c := newWithSecond()

	defer c.Stop()

	_, err := c.AddFunc(spec, p.logBak)

	logm.DebugfE("绑定定时任务")

	if err != nil {
		logm.ErrorfE("执行定时操作出错[%s]", err.Error())
		return
	}

	c.Start()

	// 信号处理
	//sig := make(chan os.Signal)
	//signal.Notify(sig, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)

	select {
	/*case noSig := <-sig:
	logm.DebugfE("接收到信号[%s]", noSig)
	return*/
	}
}

type Tables struct {
	TableName string `gorm:"size(64);column:TABLE_NAME"`
}

func (p *Program) logBak() {
	curDir, _ := os.Getwd()
	logm.InfofE("当前的工作目录[%s]", curDir)

	// 查找所有小于
	nowTime := time.Now()

	// sub 7 day
	newTime := nowTime.AddDate(0, 0, -7)

	// 组装log表的名字
	strLogName := "log_" + newTime.Format("20060102")

	// 查找所有小于strLogName的表 1、先备份数据库数据到硬盘f 2、删除数据库数据表
	var tables []*Tables
	err := GDB2.Table("TABLES").Select("TABLE_NAME").Where("TABLE_TYPE= ? and TABLE_SCHEMA = ?", "BASE TABLE", "log").Find(&tables).Error
	if err != nil {
		logm.ErrorfE("查询information_schema表出错[%s]", err.Error())
		return
	}

	for _, v := range tables {
		if v.TableName < strLogName {
			// 备份log表
			// mysqldump -uroot -hlocalhost -p123456 gamedb --hex-blob > D:\gamedb2.sql

			logm.DebugfE("备份[%s]开始", v.TableName)
			// 备份
			err, path := BackupMySqlDb1(Config.Mysql.DBHost1, "3306", Config.Mysql.DBUser, Config.Mysql.DBPasswd, Config.Mysql.DBNameA, v.TableName, "")
			if err != nil {
				logm.ErrorfE("备份表[%s] 失败[%s]", v.TableName, err.Error())
				continue
			}
			logm.DebugfE("备份[%s]结束", v.TableName)

			// 压缩
			dstFile := path + ".zip"
			logm.DebugfE("压缩[%s]开始 压缩文件[%s]", path, dstFile)

			err = Zip(path, dstFile)
			if err != nil {
				logm.ErrorfE("压缩备份表[%s] 失败[%s]", path, err.Error())
				continue
			}
			logm.DebugfE("压缩[%s]结束 压缩文件[%s]", path, dstFile)

			// 剪切文件到f盘的目录
			// 首先在f创建目录 2020/02/10/file.zip
			year, month, day, err := SplitLogTableName(v.TableName)
			if err != nil {
				year1, month1, day1 := nowTime.Date()
				year = strconv.Itoa(year1)
				month = strconv.Itoa(int(month1))
				day = strconv.Itoa(day1)
			}

			dir := Config.ZipPath + year + "\\" + month + "\\" + day
			err = CheckDir(dir)
			if err != nil {
				logm.ErrorfE("创建目录失败[%s]", err.Error())
				continue
			}

			err = ExecCmd("/C", "move /y", dstFile, dir+"\\"+dstFile)
			if err != nil {
				logm.ErrorfE("执行剪切命令失败[%s]", err.Error())
				continue
			}

			logm.DebugfE("剪切文件[%s]到[%s]成功", dstFile, dir)

			// 删除备份的原始文件
			err = os.Remove(path)
			if err != nil {
				logm.ErrorfE("删除备份的sql文件失败[%s]", err.Error())
				continue
			}

			logm.DebugfE("删除备份的sql文件成功[%s]", path)

			// 调用存储过程drop表
			err = CallDelLogProc(v.TableName)
			if err != nil {
				logm.ErrorfE("调用存储过程删除表[%s]失败[%s]", v.TableName, err.Error())
				continue
			}
			logm.DebugfE("调用存储过程删除表[%s]成功", v.TableName)
		}
	}
}

func NewDB(user, password, host, dbname, tablePrefix string) (*gorm.DB, error) {
	var err error

	//&multiStatements=true
	dia := mysql.Open(fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&multiStatements=true",
		user, password, host, dbname))
	db, err := gorm.Open(dia, &gorm.Config{NamingStrategy: schema.NamingStrategy{
		TablePrefix:   "",   // 表名前缀
		SingularTable: true, // 使用单数表名，启用该选项，此时，`Article` 的表名应该是 `article`

	},
		Logger:            logger.Default.LogMode(logger.Info),
		AllowGlobalUpdate: true,
	})

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

func StartDB() error {
	var err error
	GDB1, err = NewDB(Config.Mysql.DBUser, Config.Mysql.DBPasswd, Config.Mysql.DBHost1, Config.Mysql.DBNameA, Config.Mysql.DBTablePrefix)
	if err != nil {
		return err
	}

	GDB2, err = NewDB(Config.Mysql.DBUser, Config.Mysql.DBPasswd, Config.Mysql.DBHost2, Config.Mysql.DBNameB, Config.Mysql.DBTablePrefix)
	if err != nil {
		return err
	}

	return nil
}

func newWithSecond() *cron.Cron {
	secondParser := cron.NewParser(cron.Second | cron.Minute |
		cron.Hour | cron.Dom | cron.Month | cron.DowOptional | cron.Descriptor)
	return cron.New(cron.WithParser(secondParser), cron.WithChain())
}

// 删除指定的表
func CallDelLogProc(tblName string) error {
	str := fmt.Sprintf("call del_log('%s');", tblName)
	err := GDB1.Exec(str).Error
	if err != nil {
		return err
	}

	return nil
}

func Init() error {
	// 配置文件加载
	err := ReadConfig()
	if err != nil {
		logm.ErrorfE("读取配置文件失败[%s]", err.Error())
		return err
	}
	logm.DebugfE("读取配置文件结束")

	// 数据库初始化
	err = StartDB()
	if err != nil {
		logm.ErrorfE("初始化数据库失败[%s]", err.Error())
		return err
	}
	logm.DebugfE("初始化数据库结束")

	return nil
}

func InitFlag() {
	flag.StringVar(&Params.Dir, "dir", "", "可执行程序的目录 install的cmd下为必填的选项")
	flag.StringVar(&Params.Operator, "cmd", "", "操作的命令")
	flag.BoolVar(&Params.Daemon, "d", false, "是否后台执行 为true后台 false前台执行 其他参数无需填写")
}
