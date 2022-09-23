package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

var BitfulUrl = ""

var (
	buildTime, commitId, versionData, author, cover, url string
	help, version                                        bool
	imgUrls                                              []string
	// HttpClient 注意client 本身是连接池，不要每次请求时创建client
	dialer = &net.Dialer{
		Resolver: &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{
					Timeout: time.Duration(5000) * time.Millisecond,
				}
				return d.DialContext(ctx, "udp", "8.8.8.8:53")
			},
		},
	}
	dialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return dialer.DialContext(ctx, network, addr)
	}
	HttpClient = &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Proxy:           http.ProxyFromEnvironment,
			DialContext:     dialContext,
		},
	}
)

const (
	InfoColor    = "\033[1;34m%s\033[0m\n"
	NoticeColor  = "\033[1;36m%s\033[0m\n"
	WarningColor = "\033[1;33m%s\033[0m\n"
	ErrorColor   = "\033[1;31m%s\033[0m\n"
	DebugColor   = "\033[0;36m%s\033[0m\n"
)

func init() {
	log.SetFlags(0)
	log.Printf(InfoColor, "\n-------------------------------")
	//os.Args = append(os.Args, "~/Downloads/20201208143114085332.png", "-c", "-n", "test", "c:\\", "d:\\")
	// 美丽的身材千篇一律，有趣的灵魂万里挑一
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i-1] == "-u" || os.Args[i-1] == "--url" {
			continue
		}
		if strings.Contains(os.Args[i], "/") || strings.Contains(os.Args[i], "\\") {
			imgUrls = append(imgUrls, os.Args[i])
			os.Args = append(os.Args[:i], os.Args[(i+1):]...)
			i -= 1
		}
	}

	flag.BoolVar(&help, "h", false, "show help")
	flag.BoolVar(&help, "help", false, "show help")
	flag.BoolVar(&version, "v", false, "show version")
	flag.BoolVar(&version, "version", false, "show version")
	flag.StringVar(&cover, "c", "", "cover image")
	flag.StringVar(&cover, "cover", "", "cover image")
	flag.StringVar(&url, "u", "", "cover image name")
	flag.StringVar(&url, "url", "", "cover image name")
	flag.Parse()

	if help {
		//flag.Usage()
		log.Printf("%v%v\n",
		fmt.Printf("\033[0;34mUsage of %v:\n\n\033[1;36m -c -cover\n\tcover image name\n -h -help\n\tshow help\n -v -version\n\tshow version\033[0m\n\u001B[1;34m-------------------------------\u001B[0m\n", os.Args[0])
		os.Exit(0)
	}
	// Version
	if version {
		log.Printf("%-15v%v%-15v%v%-15v%v%-15v%v%v",
			"Version: ", fmt.Sprintf(NoticeColor, versionData),
			"BuildTime: ", fmt.Sprintf(NoticeColor, buildTime),
			"Author: ", fmt.Sprintf(NoticeColor, author),
			"CommitId: ", fmt.Sprintf(NoticeColor, commitId),
			fmt.Sprintf(InfoColor, "-------------------------------\n"))
		os.Exit(0)
	}

	if url != "" {
		generateConfig()
	}
}

func httpImg(imgUrl string) {
	// Get the data
	resp, err := http.Get(imgUrl)
	if err != nil {
		logEnd(ErrorColor, "http get img url Error: %v", err.Error())
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logEnd(ErrorColor, "http close body Error: %v", err.Error())
		}
	}(resp.Body)

	// Check Status
	if resp.StatusCode != 200 {
		logEnd(ErrorColor, "http get img url Error: %v", resp.Status)
	}

	// Read the data
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logEnd(ErrorColor, "http read body Error: %v", err.Error())
	}
	// Upload image
	var imgName string
	imgName = getImgName(imgUrl)
	UploadFile(imgName, data)
}

func fileImg(imgUrl string) {
	file, err := os.Open(imgUrl)
	if err != nil {
		logEnd(ErrorColor, "open file Error: %v", err.Error())
	}
	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			logEnd(ErrorColor, "close file Error: %v", err.Error())
		}
	}(file)
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		logEnd(ErrorColor, "read file Error: %v", err.Error())
	}
	imgName := getImgName(imgUrl)
	UploadFile(imgName, fileBytes)
	if err := os.Remove(imgUrl); err != nil {
		logEnd(ErrorColor, "remove file Error: %v", err.Error())
		return
	}
}

func getImgName(imgUrl string) (imgName string) {
	//filetype := http.DetectContentType(data)
	var imgType string
	fileNameAll := path.Base(imgUrl)
	fileSuffix := path.Ext(imgUrl)
	filePrefix := fileNameAll[0 : len(fileNameAll)-len(fileSuffix)]

	if len(filePrefix) <= 4 {
		imgType = fileSuffix
	} else {
		// png, jpg, jpeg, gif, bmp, webp, ico, tiff, svg
		if strings.Contains(fileSuffix, "png") {
			imgType = ".png"
		} else if strings.Contains(fileSuffix, "jpg") {
			imgType = ".jpg"
		} else if strings.Contains(fileSuffix, "jpeg") {
			imgType = ".jpeg"
		} else if strings.Contains(fileSuffix, "gif") {
			imgType = ".gif"
		} else if strings.Contains(fileSuffix, "bmp") {
			imgType = ".bmp"
		} else if strings.Contains(fileSuffix, "webp") {
			imgType = ".webp"
		} else if strings.Contains(fileSuffix, "ico") {
			imgType = ".ico"
		} else if strings.Contains(fileSuffix, "tiff") {
			imgType = ".tiff"
		} else if strings.Contains(fileSuffix, "svg") {
			imgType = ".svg"
		} else {
			logEnd(ErrorColor, "Unsupported image type: %v", fileSuffix)
		}
	}

	// 监测是否覆盖
	if cover != "" {
		imgName = cover + imgType
	} else {
		imgName = strings.Replace(time.Now().Format("20060102150405.00000"), ".", "", -1) + imgType
	}

	return imgName
}

func UploadFile(fileName string, file []byte) {
	body := new(bytes.Buffer)

	writer := multipart.NewWriter(body)

	formFile, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		logEnd(ErrorColor, "create form file Error: %v", err.Error())
	}

	if _, err = formFile.Write(file); err != nil {
		logEnd(ErrorColor, "write form file Error: %v", err.Error())
	}

	//for key, val := range params {
	//	_ = writer.WriteField(key, val)
	//}

	if err = writer.Close(); err != nil {
		logEnd(ErrorColor, "close writer Error: %v", err.Error())
	}
	url1 := getUrl()
	if url1 == "" {
		logEnd(ErrorColor, "get url Error: %v", err.Error())
	}
	req, err := http.NewRequest("POST", url1, body)
	if err != nil {
		logEnd(ErrorColor, "new request Error: %v", err.Error())
	}
	//req.Header.Set("Content-Type","multipart/form-data")
	req.Header.Add("Content-Type", writer.FormDataContentType())

	resp, err := HttpClient.Do(req)
	if err != nil {
		logEnd(ErrorColor, "http client do Error: %v", err.Error())
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			logEnd(ErrorColor, "http close body Error: %v", err.Error())
		}
	}(resp.Body)

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logEnd(ErrorColor, "read all body Error: %v", err.Error())
	}
	type Result struct {
		Success bool `json:"success"`
		Data    struct {
			Url  string `json:"url"`
			OUrl string `json:"o_url"`
		} `json:"data"`
	}
	var result Result
	err = json.Unmarshal(content, &result)
	if err != nil {
		logEnd(ErrorColor, "json unmarshal Error: %v", err.Error())
	}
	fmt.Println(strings.Replace(result.Data.Url, "&fmt=jpg", "&fmt=png", 1))
	// fmt.Println(result.Data.OUrl) // 原图
}

func logEnd(lv1, lv2, lv3 string) {
	log.Fatalf("%v%v\n",
		fmt.Sprintf(lv1,
			fmt.Sprintf(lv2, lv3)),
		fmt.Sprintf(InfoColor, "-------------------------------"))
}

func generateConfig() {
	dir, err := os.UserConfigDir()
	if err != nil {
		return
	}
	configFile := dir + "/" + path.Base(os.Args[0]) + ".yml"
	f, err := os.Create(configFile)
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			logEnd(ErrorColor, "Close File Error: %v", err.Error())
		}
	}(f)
	if err != nil {
		logEnd(ErrorColor, "Create File Error: %v", err.Error())
	}
	_, err = f.Write([]byte(url))
	fmt.Printf(InfoColor, fmt.Sprintf("配置文件生成成功: %v", configFile))
	os.Exit(0)
}

func getUrl() (url string) {
	if BitfulUrl != "" {
		return BitfulUrl
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		logEnd(ErrorColor, "Get UserConfigDir Error: %v", err.Error())
	}
	configFile := dir + "/" + path.Base(os.Args[0]) + ".yml"
	f, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Println("read fail", err)
		logEnd(ErrorColor, "Read File Error: %v", err.Error())
	}
	return string(f)
}

func main() {
	defer log.Printf(InfoColor, "-------------------------------\n")
	if len(imgUrls) == 0 {
		logEnd(WarningColor, "Please input the image url or file path: %v", fmt.Sprintf("%v", os.Args[1:]))
	}
	for _, imgUrl := range imgUrls {
		if strings.Contains(imgUrl, "http") {
			// 网络图片
			httpImg(imgUrl)
		} else if strings.Contains(imgUrl, "/") || strings.Contains(imgUrl, "\\") {
			// 本地图片
			fileImg(imgUrl)
		} else {
			logEnd(ErrorColor, "Unsupported image url: %v", imgUrl)
		}
	}
}
