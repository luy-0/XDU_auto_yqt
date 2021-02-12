package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	userName  = "1" // 请在此填写学号密码
	passWord  = "2"
	loginUrl  = "https://xxcapp.xidian.edu.cn/uc/wap/login/check"
	submitUrl = "https://xxcapp.xidian.edu.cn/ncov/wap/default/save"
	// 基于 Server 酱的推送服务, 如果你不知道这是什么意思请忽略
	// 不过最近好像 Server 酱有点问题，不保证可以成功获得消息
	SCKEY = ""
)

type ResData struct {
	Code int         `json:"e"`
	Msg  string      `json:"m"`
	Date interface{} `json:"d"`
}

func (d ResData) String() string {
	return fmt.Sprintf("code:{%v},msg:{%v},date: {}", d.Code, d.Msg)
}

func login() ([]*http.Cookie, ResData) {
	// 登陆
	data := make(url.Values)
	data.Add("username", userName)
	data.Add("password", passWord)
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

func submit(cookies []*http.Cookie) ResData {
	// 设置数据
	// 既然你都已经开始看 Go 版本的代码了，那抓个包应该不成问题吧？
	// 每个参数具体是啥意思就别纠结了，照着填不会把你带走的
	data := make(url.Values)
	data.Add("szgjcs", "")
	data.Add("szcs", "")
	data.Add("szgj", "")
	data.Add("zgfxdq", "0")
	data.Add("mjry", "0")
	data.Add("csmjry", "0")
	data.Add("tw", "2")
	data.Add("sfcxtz", "0")
	data.Add("sfjcbh", "0")
	data.Add("sfcxzysx", "0")
	data.Add("qksm", "")
	data.Add("sfyyjc", "0")
	data.Add("jcjgqr", "0")
	data.Add("remark", "")
	data.Add("address", "北京市海淀区*********")
	data.Add(
		"geo_api_info", "{\"type\":\"complete\",\"info\":\"SUCCESS\",\"status\":1,\"$Da\":\"jsonp_27810_\",\"position\":{\"Q\":40.***,\"R\":116.***,\"lng\":116.35291,\"lat\":40.01424},\"message\":\"Get geolocation time out.Get ipLocation success.Get address success.\",\"location_type\":\"ip\",\"accuracy\":null,\"isConverted\":true,\"addressComponent\":{\"citycode\":\"010\",\"adcode\":\"110108\",\"businessAreas\":[{\"name\":\"***\",\"id\":\"110105\",\"location\":{\"Q\":40.*****,\"R\":116.******,\"lng\":116.371292,\"lat\":40.*****}}],\"neighborhoodType\":\"\",\"neighborhood\":\"\",\"building\":\"\",\"buildingType\":\"\",\"street\":\"学清路\",\"streetNumber\":\"**\",\"country\":\"中国\",\"province\":\"北京市\",\"city\":\"\",\"district\":\"海淀区\",\"township\":\"****\"},\"formattedAddress\":\"北京市海淀区*******\",\"roads\":[],\"crosses\":[],\"pois\":[]}")
	data.Add("area", "北京市 海淀区")
	data.Add("province", "北京市")
	data.Add("city", "北京市")
	data.Add("sfzx", "0")
	data.Add("sfjcwhry", "0")
	data.Add("sfjchbry", "0")
	data.Add("sfcyglq", "0")
	data.Add("gllx", "")
	data.Add("glksrq", "")
	data.Add("jcbhlx", "")
	data.Add("jcbhrq", "")
	data.Add("ismoved", "0")
	data.Add("bztcyy", "")
	data.Add("sftjhb", "0")
	data.Add("sftjwh", "0")
	data.Add("sfjcjwry", "0")
	data.Add("jcjg", "")
	payload := data.Encode()
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

func ConnectXduMain() {
	cookies, resLogin := login()
	if resLogin.Code != 0 {
		// 登陆失败
		// 调用反馈
		onCall(resLogin, "Login")
		return
	}
	resSubmit := submit(cookies)
	if resSubmit.Code != 0 {
		// 提交失败
		onCall(resSubmit, "Submit")
	}
}

func onCall(data ResData, from string) {
	if SCKEY == "" {
		return
	}
	title := "Auto_YQT_" + from + "_" + data.Msg
	scUrl := "https://sc.ftqq.com/" + SCKEY + ".send?text=" + title
	http.Get(scUrl)
	fmt.Println(scUrl)
	fmt.Println("推送成功！")
}

func main() {
	// cloudfunction.Start(ConnectXduMain)	// 腾讯云函数打包专用
	// 关于 Goalng 在腾讯云的部署方法请参考这个链接：
	// https://cloud.tencent.com/document/product/583/18032
	ConnectXduMain() // 正常执行
	//login()

}
