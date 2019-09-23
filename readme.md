使用教程

一、初始化

	config := syncdb.Config{
		Prefix:"byd_",
		DB:"test",
		User:"root",
		Psd:"root123",
		Addr:"127.0.0.1:3306",
		ModelPath:"./model",
	}
	syncdb := syncdb.New(&config)

二、数据库同步到结构体

	syncdb.RunStruct()

三、结构体同步到数据库

	syncdb.RunDB(model.User{},model.Post{})


功能说明

	1、当数据库同步到结构体时，结构体中会记录下数据字段的所有信息，全部索引的信息
	2、当结构体同步到数据库时，如果是新建表，则会按照结构体中的信息同步字段和索引，如果是修改表，则仅仅只会增加字段或者修改字段