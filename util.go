package syncdb

import (
	"fmt"
	"os"
	"strings"
	"regexp"
)

//字符串大驼峰转下划线
func upToLower(s string) string  {
	t := ""

	runes := []rune(s)

	for _,v := range runes {
		//大写
		if v >= 65 && v <= 90 {
			t += "_"+string(v+32)
		}else{
			t += string(v)
		}
	}
	return t[1:]
}

//字符串下划线转大驼峰
func lowerToUp(s string) string {
	strs := strings.Split(s,"_")

	ss := ""

	for _,v := range strs  {
		ss += strings.Title(v)
	}

	return ss
}

//首字母小写
func lcFirst(s string) string  {
	if len(s) == 0 {
		return ""
	}

	f := s[0]

	if s[0] >= 65 && s[0] <= 90 {
		f = s[0] + 32
	}

	return string(f) + s[1:]
}

//结构体tag处理
func decodeStructTag(s string) map[string]string  {
	m := make(map[string]string)

	//正则处理字符串
	reg,err := regexp.Compile(`[a-zA-Z_]+:\"[^\"]*\"`)
	if err != nil {
		panic("解析tag失败:"+err.Error())
	}

	strs := reg.FindAllString(s,-1)
	for _,str := range strs  {

		//再次切割str 按照英文冒号
		ss := strings.SplitN(str,":",2)
		if len(ss) != 2 {
			continue
		}

		m[ss[0]] = strings.Trim(ss[1],`"`)
	}

	return m
}

//单个索引处理
func dBIndexToSql(k string,v []string) (str string) {
	//对字段进行拼接
	fields := ""
	for _,vs := range v {
		fields += ",`"+vs+"`"
	}
	fields = fields[1:]

	//主键
	if k == "primary" {
		str = "PRIMARY KEY ("+fields+")"
		return
	}

	ks := strings.Split(k,"@")
	if len(ks) != 3 {
		fmt.Println("索引字符串有误：",k)
		os.Exit(-1)
	}

	//唯一索引
	if ks[1] == "0" {
		str = "UNIQUE KEY `"+ks[0]+"` ("+fields+") USING " + ks[2]
		return
	}

	//普通索引
	str = " KEY `"+ks[0]+"` ("+fields+") USING " +  ks[2]

	return
}

//组装gorose_column
func getColumn(field map[string]interface {} ) string  {
	field_str := field["Type"].(string) + " "

	if field["Collation"] != nil {
		field_str += "collate " + field["Collation"].(string) + " "
	}

	if strings.ToLower(field["Null"].(string)) == "no" {
		field_str += "not null "
	}

	if field["Extra"].(string) != "" {
		field_str += field["Extra"].(string) + " "
	}

	if field["Default"] != nil {
		field_str += "default '"+ field["Default"].(string) +"' "
	}

	if field["Comment"].(string) != "" {
		field_str += "comment '" + field["Comment"].(string) + "'"
	}

	return field_str
}
