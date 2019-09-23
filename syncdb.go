package syncdb

import (
	"github.com/gohouse/gorose"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type Config struct {
	Prefix string		//表前缀
	DB string 			//数据库名
	User string			//数据库用户
	Psd	string 			//用户密码
	Addr string			//数据库tcp
	ModelPath string	//模型文件保存位置
}

type syncDB struct {
	Prefix string 		//表前缀
	DB string			//数据库名
	ModelPath string	//模型文件保存位置
	Engin *gorose.Engin	//gorose引擎
}



func New(config *Config) *syncDB  {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=true",config.User,config.Psd,config.Addr,config.DB)
	engine,err := gorose.Open(&gorose.Config{Driver: "mysql", Dsn: dsn})
	if err != nil {
		panic(err)
	}
	return &syncDB{
		Prefix:config.Prefix,
		ModelPath:config.ModelPath,
		DB:config.DB,
		Engin:engine,
	}
}

