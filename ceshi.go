//question ???将stdout重定向为response信息？？？
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	//生成client 参数为默认
	client := &http.Client{}

	//生成要访问的url
	url := "http://www.baidu.com"

	//提交请求
	reqest, err := http.NewRequest("GET", url, nil)

	if err != nil {
		panic(err)
	}

	//处理返回结果
	response, _ := client.Do(reqest)
	//将结果定位到标准输出 也可以直接打印出来 或者定位到其他地方进行相应的处理
	stdout := os.Stdout
	_, err = io.Copy(stdout, response.Body)

	//返回的状态码
	status := response.StatusCode

	fmt.Println(status)

}
