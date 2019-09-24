package syncdb

import (
	"fmt"
	"reflect"
	"sync"
)

func (syncdb *syncDB)RunDB(st ...interface{})  {
	st_len := len(st)

	if st_len == 0 {
		fmt.Println("没有输入需要同步的表")
		return
	}

	var wg sync.WaitGroup

	wg.Add(st_len)

	for _,v := range st {
		go func(v interface{}) {
			syncdb.doDB(v)
			wg.Done()
		}(v)
	}

	wg.Wait()

}

func (syncdb *syncDB)doDB(u interface{})  {
	db := syncdb.Engin.NewSession()

	fk := reflect.TypeOf(u)

	if fk.Kind().String() == "ptr" {
		fk = fk.Elem()
	}

	if fk.Kind().String() != "struct" {
		fmt.Println(upToLower(fk.Name()),"不是结构体")
		return
	}

	//获得表名称
	table_name := syncdb.Prefix + upToLower(fk.Name())

	//查看表是否存在
	res,err:=db.Query("show tables like '"+table_name+"' ")
	if err != nil {
		fmt.Println("查看表存在出错：",err)
		return
	}

	//定义表的map
	table_array := make([]string,len(res))
	table_map := make(map[string]map[string]string)
	//记录表的索引
	indexs := make(map[string][]string)

	//获取结构体的所有tag数据
	for i:=0;i<fk.NumField();i++{
		filed := fk.Field(i)

		//处理字段
		m := decodeStructTag(string(filed.Tag))
		table_map[m["gorose"]] = m

		table_array = append(table_array,m["gorose"])

		//处理索引
		indexStr,ok := m["gorose_key"]
		if ok && len(indexStr) > 0 {
			//给索引名添加新的字段
			indexs[indexStr] = append(indexs[indexStr],m["gorose"])
		}

	}

	if len(res) == 0 {
		//表不存在  创建表
		tsql := "create table `" + table_name + "` ("


		for _,key := range table_array {
			v := table_map[key]
			tsql += "`" + v["gorose"] + "` " + v["gorose_column"] + ","
		}

		//加入索引
		for k,v := range indexs {
			ss := dBIndexToSql(k, v)

			if ss == "" {
				continue
			}

			tsql += ss + ","
		}

		tsql = tsql[:len(tsql)-1] + ")"

		_,err:=db.Execute(tsql)

		if err != nil {
			fmt.Println("表",table_name,"导入失败：",err)
			return
		}

		fmt.Println("表",table_name,"导入成功")

	}else{
		//有表  判断它是否发生改变  增加的字段需要同步进去

		//获取该表所有字段信息
		res,err := db.Query("SHOW FULL COLUMNS FROM " + table_name)

		if err != nil {
			fmt.Println("查看表全字段信息出错：",err)
			return
		}

		db_map := make(map[string]map[string]interface{})
		//字段名  --  数据map对象
		for _,v := range res {
			db_map[v["Field"].(string)] = v
		}

		sqls := make([]string,0,len(table_map))

		//获取到新增的字段
		for k,v := range table_map {
			//判断 结构体 的字段  是否存在于 表字段中
			field,ok := db_map[k]

			if !ok {
				//不存在 则新增
				sql := "ALTER TABLE `"+table_name+"` ADD COLUMN `" + v["gorose"] +"` " + v["gorose_column"]
				sqls = append(sqls,sql)
				continue
			}

			//存在  则判断是否需要修改
			//判断字段信息是否发生变化 如果有  则对字段进行修改
			//如果发生了变化 则说明有问题
			if v["gorose_column"] != getColumn(field) {
				sql := "ALTER TABLE `"+table_name+"` MODIFY COLUMN " + v["gorose"] +"` " + v["gorose_column"]
				sqls = append(sqls,sql)
			}

		}

		if len(sqls) == 0{
			fmt.Println("表",table_name,"没有字段变化")
			return
		}

		//执行sql
		db.Begin()
		for _,str := range sqls {
			_,err:=db.Execute(str)
			if err != nil {
				fmt.Println("表",table_name,"修改失败：",str,err)
				db.Rollback()
				return
			}
		}
		db.Commit()
		fmt.Println("表",table_name,"修改成功")
	}
}