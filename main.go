package main

import (
	"bufio"
	"compress/flate"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/dsnet/compress/brotli"
)

const burl = "http://share.dmhy.org"

const curl = "http://share.dmhy.org/topics/list/page/2?keyword=%E5%A6%82%E6%9E%9C%E6%9C%89%E5%A6%B9%E5%A6%B9"

const MAX_PAGE_NUM = 3

type FUrlInfo struct {
	dtype      string
	authorName string
	title      string
	url        string
	size       string
}

type DirInfo struct {
	dirName string
	dmList  []string
	thisDM  string
}

func (d *DirInfo) dealDir() {
	dir := "./data/" + d.dirName
	if _checkFileExist(dir) {
		fmt.Println(dir, "已存在")
	} else {
		os.Mkdir(dir, os.ModePerm)
		d.dirName = dir
		d.dealDM()
	}
}

func (d *DirInfo) dealDM() {
	for _, dmName := range d.dmList {
		d.thisDM = dmName
		dmNameEnCode := url.QueryEscape(dmName)
		for i := 1; i <= MAX_PAGE_NUM; i++ {
			page := fmt.Sprintf("page/%v", i)
			url := "https://share.dmhy.org/topics/list/" + page + "?keyword=" + dmNameEnCode
			d.sendGetReq(url)
		}
	}
}

var fileList []string

func main() {
	//_sendGetReq(curl)
	//_dealFPageHtml(ret)
	//var f FUrlInfo
	//_readList(name)
	//url := "http://share.dmhy.org/topics/view/476365_10_07_BIG5_1080P_MP4.html"
	//_sendSonGetReq(&FUrlInfo{url: url})
	slc := _getDMName()
	for _, dirinfo := range slc {
		dirinfo.dealDir()

	}
}

func _getFileList() {
	dir, err := ioutil.ReadDir("./list")
	if err != nil {
		fmt.Println("读取目录失败")
	}
	for _, f := range dir {
		if f.IsDir() {
			continue
		}
		fileList = append(fileList, f.Name())
	}
}

func _readFile(file string) []string {
	dir := "./list/"
	file = dir + file
	var f *os.File
	var dmNameList []string
	defer f.Close()
	f, err := os.Open(file)
	if err != nil {
		fmt.Println()
	}
	fl := bufio.NewReader(f)
	for {
		name, _, err := fl.ReadLine()
		if err == io.EOF {
			break
		}
		dmNameList = append(dmNameList, string(name))
	}
	return dmNameList
}

func _getDMName() []DirInfo {
	_getFileList()
	var v []DirInfo
	if len(fileList) != 0 {
		for _, file := range fileList {
			var dirInfo DirInfo
			nameList := _readFile(file)
			dirInfo.dirName = file
			for _, dmName := range nameList {
				dirInfo.dmList = append(dirInfo.dmList, dmName)
			}
			v = append(v, dirInfo)
		}
	}
	return v
}

func (d *DirInfo) sendGetReq(url string) {
	fmt.Println(url)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Host", "share.dmhy.org")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:55.0) Gecko/20100101 Firefox/55.0")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.8,en-US;q=0.5,en;q=0.3")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Referer", "https://share.dmhy.org/")
	req.Header.Add("Cookie", "__cfduid=dce9622a13b31e6fee4113a6bb98d281d1514368352; Hm_lvt_e4918ccc327a268ee93dac21d5a7d53c=1514368352,1514864052,1514886699; HstCfa3801674=1514368352885; HstCla3801674=1514886710356; HstCmu3801674=1514368352885; HstPn3801674=22; HstPt3801674=39; HstCnv3801674=9; HstCns3801674=15; __dtsu=2DE7B66B6B6D435ABB1D0D2A02676A23; Hm_lpvt_e4918ccc327a268ee93dac21d5a7d53c=1514886711")
	req.Header.Add("Upgrade-Insecure-Requests", "1")
	req.Header.Add("If-Modified-Since", "Tue, 02 Jan 2018 09:54:56 GMT")
	req.Header.Add("Cache-Control", "max-age=0")
	if err != nil {
		fmt.Println("请求失败")
		panic(err)
	}
	resq, err := client.Do(req)
	if err != nil {
		fmt.Println("发送失败")
	}
	defer resq.Body.Close()

	//reader, _ := gzip.NewReader(resq.Body)
	entype := resq.Header.Get("content-encoding")
	switch entype {
	case "br":
		reader, _ := brotli.NewReader(resq.Body, nil)
		d.dealFPageHtml(reader)
	case "deflate":
		reader := flate.NewReader(resq.Body)
		d.dealFPageHtml(reader)
	case "gzip":
		reader, _ := gzip.NewReader(resq.Body)
		d.dealFPageHtml(reader)
	}
	//ret, err := ioutil.ReadAll(reader)
	//fmt.Println(string(ret))
	//log := "./log/log.txt"
	//f, _ := os.Create(log)
	//_, err = io.WriteString(f, string(ret))
	//if err != nil {
	//	fmt.Println("写入失败")
	//}
}

func (d *DirInfo) sendSonGetReq(f *FUrlInfo) {
	time.Sleep(5 * time.Second)
	url := f.url
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Host", "share.dmhy.org")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:55.0) Gecko/20100101 Firefox/55.0")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.8,en-US;q=0.5,en;q=0.3")
	req.Header.Add("Accept-Encoding", "gzip, deflate")
	req.Header.Add("Cookie", "__cfduid=dce9622a13b31e6fee4113a6bb98d281d1514368352; Hm_lvt_e4918ccc327a268ee93dac21d5a7d53c=1514368352,1514864052; HstCfa3801674=1514368352885; HstCla3801674=1514877987425; HstCmu3801674=1514368352885; HstPn3801674=14; HstPt3801674=31; HstCnv3801674=9; HstCns3801674=12; __dtsu=2DE7B66B6B6D435ABB1D0D2A02676A23; Hm_lpvt_e4918ccc327a268ee93dac21d5a7d53c=1514877987")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Upgrade-Insecure-Requests", "1")
	req.Header.Add("If-Modified-Since", "Tue, 02 Jan 2018 07:29:33 GMT")
	req.Header.Add("Cache-Control", "max-age=0")
	if err != nil {
		fmt.Println("请求失败")
		panic(err)
	}
	resq, err := client.Do(req)
	if err != nil {
		fmt.Println("发送失败")
	}
	defer resq.Body.Close()

	reader, _ := gzip.NewReader(resq.Body)
	d.dealSonPageHtml(reader, f)

}

func (d *DirInfo) dealFPageHtml(r io.Reader) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		fmt.Println("html解析出错")
	}
	//ret, _ := doc.Find("div.clear").Find("tr").Find("td.title").Find("span.tag").Find("a").Attr("href")
	//fmt.Println(ret.Text())
	doc.Find("div.clear").Find("tbody").Find("tr").Each(func(index int, sel *goquery.Selection) {
		dtype := sel.Find("td").Eq(1).Find("font").Text()
		fmt.Println(dtype)
		if dtype == "季度全集" {
			authorName := sel.Find("td.title").Find("span.tag").Find("a").Text()
			fhtml := sel.Find("td.title").Find("[target='_blank']")
			size := sel.Find("[nowrap='nowrap']").Eq(1).Text()
			surl, _ := fhtml.Attr("href")
			title := fhtml.Text()
			surl = burl + surl

			var finfo FUrlInfo
			finfo.dtype = dtype
			finfo.authorName = authorName
			finfo.title = title
			finfo.url = surl
			finfo.size = size
			//d.sendSonGetReq(&finfo)
			fmt.Println(finfo)
			//fmt.Println("*************")
		}
	})
	//_saveInFile(str)
}

func (d *DirInfo) dealSonPageHtml(r io.Reader, f *FUrlInfo) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		fmt.Println("html解析出错")
		return
	}
	torrUrl, _ := doc.Find("div#tabs-1").Find("a").Eq(0).Attr("href")
	magStr, _ := doc.Find("div#tabs-1").Find("a").Eq(1).Attr("href")
	//	fmt.Println(torrUrl, magStr)
	d.makeSonDir()
	d.downLoadTorr(torrUrl, f)
	d.saveMagStr(magStr, f)
}

func (d *DirInfo) makeSonDir() {
	dir := d.dirName + d.thisDM
	if !_checkFileExist(dir) {
		os.Mkdir(dir, os.ModePerm)
	}
}

func (d *DirInfo) downLoadTorr(torrUrl string, f *FUrlInfo) {
	torrUrl = "http:" + torrUrl
	res, err := http.Get(torrUrl)
	defer res.Body.Close()
	file := d.dirName + "/" + d.thisDM + "/" + f.title + "+" + f.size + ".torrent"
	if err != nil {
		fmt.Println("下载失败")
		return
	}
	if _checkFileExist(file) {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		num := r.Intn(100)
		file = d.dirName + "/" + d.thisDM + "/" + f.title + "+" + f.size + string(num) + ".torrent"
	}
	fp, _ := os.Create(file)
	io.Copy(fp, res.Body)
}

func (d *DirInfo) saveMagStr(str string, f *FUrlInfo) {
	file := d.dirName + "/" + d.thisDM + "/" + f.title + "+" + f.size + ".txt"
	if _checkFileExist(file) {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		num := r.Intn(100)
		file = d.dirName + "/" + d.thisDM + "/" + f.title + "+" + f.size + string(num) + ".txt"
	}
	fp, _ := os.Create(file)
	_, err := io.WriteString(fp, str)
	if err != nil {
		fmt.Println("写入mag失败")
	}
}

func _saveInFile(slc []string) {
	//fmt.Println(file)
	log := "./log/log.txt"
	time := time.Now().Format("2006-01-02_15:04:05")
	fileName := log + "." + time
	fmt.Println(fileName)
	var f *os.File
	defer f.Close()
	//	if _checkFileExist(log) {
	//		f, _ = os.OpenFile(fileName, os.O_APPEND, 0666)
	//	} else {
	f, _ = os.Create(fileName)
	//	}
	for _, file := range slc {
		_, err := io.WriteString(f, file)
		if err != nil {
			fmt.Println("写入文件错误")
		}
	}
}

func _checkFileExist(f string) bool {
	var exist = true
	if _, err := os.Stat(f); os.IsNotExist(err) {
		exist = false
	}
	return exist
}
