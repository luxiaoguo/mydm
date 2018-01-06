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
	"strings"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"github.com/dsnet/compress/brotli"
)

const burl = "https://share.dmhy.org"

const MAX_PAGE_NUM = 5

type FUrlInfo struct {
	dtype      string
	authorName string
	title      string
	url        string
	size       string
	simple     string
}

type DirInfo struct {
	dirName string
	dmList  []string
	thisDM  string
	isFind  bool
	srcNum  uint
}

func (d *DirInfo) dealDir() {
	dir := "./data/" + d.dirName
	if _checkFileExist(dir) {
		log("已存在", dir)
	} else {
		os.Mkdir(dir, os.ModePerm)
		d.dirName = dir
		d.dealDM()
	}
}

func (d *DirInfo) dealDM() {
	for _, dmName := range d.dmList {
		d.isFind = false
		d.srcNum = 0
		nameSlc := strings.Split(dmName, "|")
		d.thisDM = nameSlc[0]
		for _, sName := range nameSlc {
			dmNameEnCode := url.QueryEscape(sName)
			for i := 1; i <= MAX_PAGE_NUM; i++ {
				page := fmt.Sprintf("page/%v", i)
				url := "https://share.dmhy.org/topics/list/" + page + "?keyword=" + dmNameEnCode
				d.sendGetReq(url)
			}
		}
		if !d.isFind {
			unfindList(d.thisDM)
		} else {
			finishList(d.thisDM, d.srcNum)
		}
	}
}

var fileList []string

func main() {
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
			dirInfo.dirName = _delFix(file)
			for _, dmName := range nameList {
				dirInfo.dmList = append(dirInfo.dmList, dmName)
			}
			v = append(v, dirInfo)
		}
	}
	return v
}

func (d *DirInfo) sendGetReq(url string) {
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
		fmt.Println("创建请求失败")
		log("请求失败", d.thisDM)
		return
	}
	resq, err := client.Do(req)
	if err != nil {
		fmt.Println("发送失败")
		log("发送请求失败", d.thisDM)
		return
	}
	defer resq.Body.Close()

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
}

func (d *DirInfo) sendSonGetReq(f *FUrlInfo) {
	time.Sleep(3 * time.Second)
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
		log("请求失败", d.thisDM)
		return
	}
	resq, err := client.Do(req)
	if err != nil {
		fmt.Println("发送失败")
		log("发送请求失败", d.thisDM)
		return
	}
	defer resq.Body.Close()

	reader, _ := gzip.NewReader(resq.Body)
	d.dealSonPageHtml(reader, f)

}

func (d *DirInfo) dealFPageHtml(r io.Reader) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		log("html解析出错，应该是相应结果为空", d.thisDM)
		return
	}
	doc.Find("div.clear").Find("tbody").Find("tr").Each(func(index int, sel *goquery.Selection) {
		dtype := sel.Find("td").Eq(1).Find("font").Text()
		if dtype == "季度全集" {
			d.isFind = true
			d.srcNum++
			authorName := sel.Find("td.title").Find("span.tag").Find("a").Text()
			fhtml := sel.Find("td.title").Find("[target='_blank']")
			size := sel.Find("[nowrap='nowrap']").Eq(1).Text()
			surl, aru := fhtml.Attr("href")
			title := _delStrSpa(fhtml.Text())
			if !aru {
				log("未获取到url", title)
				return
			}
			surl = burl + surl

			var finfo FUrlInfo
			finfo.dtype = dtype
			finfo.authorName = authorName
			finfo.title = title
			finfo.url = surl
			finfo.size = size
			//fmt.Println(title)
			//fmt.Println([]byte(title))
			d.sendSonGetReq(&finfo)
		}
	})
	//_saveInFile(str)
}

func (d *DirInfo) dealSonPageHtml(r io.Reader, f *FUrlInfo) {
	time.Sleep(3 * time.Second)
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		fmt.Println("html解析出错")
		log("html解析出错，应该是相应结果为空", d.thisDM)
		return
	}
	torrUrl, _ := doc.Find("div#tabs-1").Find("a").Eq(0).Attr("href")
	magStr, _ := doc.Find("div#tabs-1").Find("a").Eq(1).Attr("href")
	simple, _ := doc.Find(".topic-nfo").Html()
	d.makeSonDir()
	d.downLoadTorr(torrUrl, f)
	d.saveMagStr(magStr, f)
	d.saveSimStr(simple, f)
}

func (d *DirInfo) makeSonDir() {
	dir := d.dirName + "/" + d.thisDM
	if !_checkFileExist(dir) {
		os.Mkdir(dir, os.ModePerm)
	}
}

func (d *DirInfo) downLoadTorr(torrUrl string, f *FUrlInfo) {
	torrUrl = "https:" + torrUrl
	res, err := http.Get(torrUrl)
	defer res.Body.Close()
	file := d.dirName + "/" + d.thisDM + "/" + f.title + "+" + f.size + ".torrent"
	if err != nil {
		fmt.Println("下载失败")
		log("下载torr文件失败,尝试第二次下载", f.title+f.size)
		res, err = http.Get(torrUrl)
		if err != nil {
			log("第二次下载失败", f.title+f.size)
			return
		}
	}
	if _checkFileExist(file) {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		num := r.Intn(100)
		file = d.dirName + "/" + d.thisDM + "/" + f.title + "+" + f.size + string(num) + ".torrent"
	}
	//fmt.Println(file)
	fp, err := os.Create(file)
	defer fp.Close()
	if err != nil {
		fmt.Println("torr创建失败")
		log("创建torr文件失败", f.title+f.size)
		return
	}
	io.Copy(fp, res.Body)
	log("torr文件下载成功", f.title+f.size)
}

func (d *DirInfo) saveMagStr(str string, f *FUrlInfo) {
	file := d.dirName + "/" + d.thisDM + "/" + f.title + "+" + f.size + "_" + "mag" + ".txt"
	if _checkFileExist(file) {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		num := r.Intn(100)
		file = d.dirName + "/" + d.thisDM + "/" + f.title + "+" + f.size + string(num) + "_" + "mag" + ".txt"
	}
	fmt.Println(file)
	fp, err := os.Create(file)
	if err != nil {
		fmt.Println("创建文件失败")
		log("创建mag文件失败", f.title+f.size)
		return
	}
	defer fp.Close()
	_, err = io.WriteString(fp, str)
	if err != nil {
		fmt.Println("写入mag失败")
		log("写入mag失败", f.title+f.size)
		return
	}
	log("mag文件写入成功", f.title+f.size)
}

func (d *DirInfo) saveSimStr(str string, f *FUrlInfo) {
	file := d.dirName + "/" + d.thisDM + "/" + f.title + "+" + f.size + "_" + "simple" + ".txt"
	if _checkFileExist(file) {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		num := r.Intn(100)
		file = d.dirName + "/" + d.thisDM + "/" + f.title + "+" + f.size + string(num) + "_" + "simple" + ".txt"
	}
	fmt.Println(file)
	fp, err := os.Create(file)
	if err != nil {
		fmt.Println("创建文件失败")
		log("创建简介文件失败", f.title+f.size)
		return
	}
	defer fp.Close()
	_, err = io.WriteString(fp, str)
	if err != nil {
		fmt.Println("写入简介失败")
		log("写入简介文件失败", f.title+f.size)
		return
	}
	log("写入简介文件成功", f.title+f.size)
}

//func (d *DirInfo) getNextPage(doc *goquery.Document) string{

//}

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

func _delStrSpa(str string) string {
	strSlc := strings.FieldsFunc(str, unicode.IsSpace)
	var nstr string
	for _, v := range strSlc {
		nstr = nstr + v
	}
	nstr = strings.Replace(str, ".", "_", -1)
	nstr = strings.Replace(str, "/", "_", -1)
	bslc := []byte(nstr)
	bslc = bslc[5:]
	return string(bslc)
}

func log(state string, name string) {
	time := time.Now().Format("2006-01-02 15:04:05")
	str := time + " " + "[" + name + "]" + " " + state + "\n"
	var f *os.File
	if _checkFileExist("log.txt") {
		f, _ = os.OpenFile("log.txt", os.O_WRONLY|os.O_APPEND, 0644)
	} else {
		f, _ = os.Create("log.txt")
	}
	defer f.Close()
	//	if err != nil {
	//		fmt.Println("open log.txt err ")
	//	}
	io.WriteString(f, str)
}

func _delFix(str string) string {
	slc := strings.Split(str, ".")
	return slc[0]
}

func unfindList(name string) {
	str := name + "未找到资源"
	var f *os.File
	if _checkFileExist("unfind.txt") {
		f, _ = os.OpenFile("unfind.txt", os.O_WRONLY|os.O_APPEND, 0644)
	} else {
		f, _ = os.Create("unfind.txt")
	}
	defer f.Close()
	str = str + "\n"
	io.WriteString(f, str)
}

func finishList(name string, num uint) {
	str := fmt.Sprintf("%v:%v获取到%v个资源", time.Now().Format("2006-01-02 15:04:05"), name, num)
	var f *os.File
	if _checkFileExist("finish.txt") {
		f, _ = os.OpenFile("finish.txt", os.O_WRONLY|os.O_APPEND, 0644)
	} else {
		f, _ = os.Create("finish.txt")
	}
	defer f.Close()
	str = str + "\n"
	io.WriteString(f, str)
}
