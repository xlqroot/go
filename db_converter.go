package syncdb

import (
	"fmt"
	"sync"
	"strings"
	"strconv"
	"os"
	"os/exec"
)

//写入所有表到struct中
var typeForMysqlToGo = map[string]string{
	"int":                "int",
	"integer":            "int",
	"tinyint":            "int",
	"smallint":           "int",
	"mediumint":          "int",
	"bigint":             "int64",
	"int unsigned":       "int",
	"integer unsigned":   "int",
	"tinyint unsigned":   "int",
	"smallint unsigned":  "int",
	"mediumint unsigned": "int",
	"bigint unsigned":    "int64",
	"bit":                "int",
	"bool":               "bool",
	"enum":               "string",
	"set":                "string",
	"varchar":            "string",
	"char":               "string",
	"tinytext":           "string",
	"mediumtext":         "string",
	"text":               "string",
	"longtext":           "string",
	"blob":               "string",
	"tinyblob":           "string",
	"mediumblob":         "string",
	"longblob":           "string",
	"date":               "string", // time.Time or string
	"datetime":           "string", // time.Time or string
	"timestamp":          "string", // time.Time or string
	"time":               "string", // time.Time or string
	"float":              "float64",
	"double":             "float64",
	"decimal":            "float64",
	"binary":             "string",
	"varbinary":          "string",
}

func (syncdb *syncDB)RunStruct()  {
	db := syncdb.Engin.NewSession()

	//获取所有表名
	res,err := db.Query("show tables" )

	if err != nil {
		fmt.Println("获取所有表名称失败:",err)
		return
	}

	var wg sync.WaitGroup

	wg.Add(len(res))

	for _,v := range res  {
		s := v["Tables_in_"+syncdb.DB].(string)
		go func(ss string) {
			syncdb.doStruct(ss)
			wg.Done()
		}(s)
	}
	wg.Wait()

}

//根据表名生成 结构体
func (syncdb *syncDB)doStruct(s string)  {
	if !strings.HasPrefix(s,syncdb.Prefix) {
		fmt.Println("表",s,"不包含前缀：",syncdb.Prefix)
		return
	}

	db := syncdb.Engin.NewSession()
	//获取表的所有信息
	res,err:=db.Query("show full columns from " + s)

	if err != nil {
		fmt.Println("表",s,"获取字段失败：",err)
		return
	}

	//获取表的所有索引信息
	ires,err := db.Query("show keys from " + s)
	if err != nil {
		fmt.Println("表",s,"获取索引信息失败：",err)
		return
	}

	indexs := make(map[string]string)

	for _,m := range ires {
		key := m["Column_name"].(string)
		if m["Key_name"].(string) == "PRIMARY" {
			indexs[key] = "primary"
			continue
		}

		indexs[key] = strings.ToLower(m["Key_name"].(string)) + "@" + strconv.FormatInt(m["Non_unique"].(int64),10) + "@" + strings.ToLower(m["Index_type"].(string))
	}


	str := "package model \n\n"

	str += "type "+lowerToUp(strings.TrimPrefix(s ,syncdb.Prefix))+" struct { \n\n"

	for _,m := range res {
		key := lowerToUp(m["Field"].(string))  //结构体的字段

		str += key + " "

		ftype := ""

		ss := strings.Split(m["Type"].(string)," ")

		for _,v := range ss {
			index := strings.Index(v,"(")

			if index >= 0 {
				ftype += v[:strings.Index(v,"(")]+ " " + v[strings.Index(v,")")+1:]
			}else{
				ftype += v
			}
		}

		ftype = strings.TrimSpace(ftype)

		str += typeForMysqlToGo[ftype] + " "

		str += "`"

		str += `gorose:"`+m["Field"].(string)+`" `

		str += `gorose_column:"`+ getColumn(m)+`" `

		//写入索引
		str += ` gorose_key:"`+ indexs[m["Field"].(string)] +`"`

		//写入json
		str += ` json:"`+ lcFirst(key) +`"`

		//tag 结束
		str += "`"

		//写入备注
		if m["Comment"].(string) != "" {
			str += "//" + m["Comment"].(string)
		}

		//一个字段完结
		str += "\n"

	}

	str += "}\n\n"

	//开始写入文件
	dir := syncdb.ModelPath

	if _,err := os.Stat(dir);err !=nil {
		//目录不存在  创建目录
		os.Mkdir(dir,0777)
		os.Chmod(dir, 0777)
	}

	filename := strings.Replace(s, syncdb.Prefix, "", -1)+".go"
	if strings.HasSuffix(dir,"/") {
		filename = dir + filename
	}else{
		filename = dir + "/" + filename
	}

	f,err := os.OpenFile(filename,os.O_CREATE|os.O_RDWR|os.O_TRUNC ,0666)
	if err != nil {
		fmt.Println("表",s,"创建模型文件失败",err)
		return
	}
	defer f.Close()

	//清空文件
	f.Truncate(-1)

	n,err:=f.WriteString(str)

	if err != nil {
		fmt.Println("表",s,"写入模型文件失败",err)
		return
	}

	if n <= 0 {
		fmt.Println("表",s,"写入模型文件没有成功",err)
		return
	}

	cmd := exec.Command("gofmt","-w",filename)
	cmd.Run()

	fmt.Println("表",s,"写入模型文件成功")

}