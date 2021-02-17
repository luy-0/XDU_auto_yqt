package main

import (
	"encoding/json"
	"fmt"
	"github.com/tencentyun/scf-go-lib/cloudfunction"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

const (
	loginUrl   = "https://xxcapp.xidian.edu.cn/uc/wap/login/check"
	requestURL = "https://xxcapp.xidian.edu.cn/ncov/wap/default/index"
	submitUrl  = "https://xxcapp.xidian.edu.cn/ncov/wap/default/save"
)

type ResData struct {
	// 定义了疫情通的返回值
	Code int         `json:"e"`
	Msg  string      `json:"m"`
	Date interface{} `json:"d"`
}

func (d ResData) String() string {
	return fmt.Sprintf("code:{%v},msg:{%v},date: {}", d.Code, d.Msg)
}

func login() ([]*http.Cookie, ResData) {
	// 登陆， 从环境变量中直接获得学号密码
	// 请提前配置好环境变量
	data := make(url.Values)
	data.Add("username", os.Getenv("username"))
	data.Add("password", os.Getenv("password"))
	payload := data.Encode()

	res, _ := http.DefaultClient.Post(loginUrl, "application/x-www-form-urlencoded", strings.NewReader(payload))
	defer func() { _ = res.Body.Close() }()

	// 打印登陆状态
	content, _ := ioutil.ReadAll(res.Body)
	var resData ResData
	json.Unmarshal(content, &resData)
	fmt.Printf("Log in : %s\n\n", resData)

	// 返回 cookies 与登陆状态
	return res.Cookies(), resData
}

func submitPlus(cookies []*http.Cookie) ResData {
	// 填充数据
	// 获取前一天的填报数据
	payload := generateData(cookies)
	req, _ := http.NewRequest(http.MethodPost, submitUrl, strings.NewReader(payload))

	// 设置 cookies
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// submit
	res, _ := http.DefaultClient.Do(req)
	defer func() { _ = res.Body.Close() }()

	// 打印状态
	content, _ := ioutil.ReadAll(res.Body)
	var resData ResData
	resData.Code = -1
	json.Unmarshal(content, &resData)
	fmt.Printf("Submit : %s\n\n", resData)

	return resData
}

func generateData(cookies []*http.Cookie) string {
	// 获取 html 文件
	req, _ := http.NewRequest(http.MethodGet, requestURL, nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	res, _ := http.DefaultClient.Do(req)
	defer func() { _ = res.Body.Close() }()
	content, _ := ioutil.ReadAll(res.Body)

	// 解析出 oldInfo
	reg1 := regexp.MustCompile(`oldInfo: (.*),`)
	if reg1 == nil { //解释失败，返回nil
		fmt.Println("regexp err")
		panic("解释器失败")
	}
	//根据规则提取关键信息
	result := reg1.FindAllStringSubmatch(string(content), -1)

	jsonStr := result[0][1]

	data := make(url.Values)
	for s, result := range gjson.Parse(jsonStr).Map() {
		data.Add(s, result.String())
	}
	data.Set("tw", "1")
	data.Set("ismoved", "0")
	data.Set("mjry", "0")
	data.Set("csmjry", "0")
	data.Set("zgfxdq", "0")

	payload := data.Encode()

	return payload
}

func ConnectXduMain() {
	cookies, resLogin := login()
	if resLogin.Code != 0 {
		onCall(resLogin, "Login")
		return
	}
	resSubmit := submitPlus(cookies)
	if resSubmit.Code != 0 {
		onCall(resSubmit, "Submit")
	}
}

func onCall(data ResData, from string) {
	title := "Auto_YQT_" + from + "_" + data.Msg
	fmt.Println(title)
}

func main() {
	// 本地测试去掉下面四行注释
	//var event DefineEvent
	//event.username=""
	//event.Password=""
	//ConnectXduMain(nil,event)

	// 云函数打包专用
	cloudfunction.Start(ConnectXduMain) // 腾讯云函数打包专用
}
