package main

import (
	"goSendMail/email"
	"log"
	"os"
	"strings"
)

func main() {
	//初始化配置
	err := email.LoadEmailConfig("./conf", "conf", "yaml")
	if err != nil {
		panic("load conf failed")
		return
	}
	email.StartEmail()
	//发送普通邮件
	email.Send(email.NewSender("这是普通邮件", "hello iamzcr!", "randzcr@gmail.com"))
	//发送模板邮件
	email.Send(email.NewHTMLSender("这是模板邮件", email.ParseHHTML(GetRootPath()+"/html/email.html", struct{ Name string }{Name: "iamzcr"}), "randzcr@gmail.com"))

}

func GetRootPath() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}
