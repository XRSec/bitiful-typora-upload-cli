package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/Bios-Marcel/wastebasket"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

// S3 配置及全局变量
var (
	Endpoint    = "s3.bitiful.net" // S3 终端节点
	Region      = "cn-east-1"      // S3 区域
	SVC         *s3.S3             // S3 客户端
	file        string             // 文件路径或URL
	name        string             // 指定上传后的文件名
	verbose     bool               // 是否显示详细日志
	HttpClient  *http.Client       // 全局 HTTP 客户端
	versionData string             // 版本信息
	fileName    string             // 实际上传到S3的文件名
	imageType   bool               // 是否为本地文件（true=本地，false=URL）
	err         error              // 全局错误变量
)

// 初始化配置和HTTP客户端
func init() {
	dir, err := os.UserConfigDir()
	if err != nil {
		return
	}
	userPath := strings.ReplaceAll(dir, "\\", "/")
	configFile := path.Join(userPath, "bitiful.yml")
	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}
	if viper.GetString("Endpoint") != "" {
		Endpoint = viper.GetString("Endpoint")
	}
	if viper.GetString("Region") != "" {
		Region = viper.GetString("Region")
	}

	// 初始化全局 HTTP 客户端，避免频繁创建
	dialer := &net.Dialer{
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
	dialContext := func(ctx context.Context, network, addr string) (net.Conn, error) {
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
}

// 主命令定义，支持批量文件/URL上传
var rootCmd = &cobra.Command{
	Short:   "Typora 上传图片到 AWS",
	Long:    "优雅地使用 Typora 上传图片到 AWS 存储桶",
	Version: versionData,

	Run: func(cmd *cobra.Command, args []string) {
		// 如果没有图片路径参数，则用 -f 指定的文件
		if args == nil || len(args) == 0 {
			args = append(args, file)
		}
		log.Debugf("args: %v", args)
		log.Debugf("file: %v, name: %v", file, name)
		for _, v := range args {
			var data []byte
			log.Debugf("处理文件/URL: %s", v)
			getImgName(v)
			log.Debugf("最终上传文件名: %s", fileName)
			imageType = checkImg(v)
			log.Debugf("文件类型判断: %v (true=本地文件, false=URL)", imageType)
			// 判断是本地文件还是URL，分别处理
			if !imageType {
				log.Debugf("尝试下载图片: %s", v)
				data, err = getImageUrl(v)
				if err != nil {
					log.Fatalln(err)
					continue
				}
			} else {
				log.Debugf("尝试读取本地图片: %s", v)
				data, err = getImageFile(v)
				if err != nil {
					log.Fatalln(err)
					continue
				}
			}

			log.Debugf("开始上传到 S3: %s", fileName)
			_, err := UploadObject(data)
			if err != nil {
				log.Fatalln(err)
				continue
			}
			log.Debugf("上传成功: %s", fileName)

			if imageType {
				// 上传图片后删除原文件
				if err := wastebasket.Trash(v); err != nil {
					log.Errorf("删除文件失败: %v", err)
					return
				}
				log.Debugf("移动图片到垃圾篓成功: %s", fileName)
			}
			fmt.Printf("https://%v.%v%v%v?fmt=webp&q=48&w=800\n", viper.GetString("BucketName"), Endpoint, viper.GetString("Path"), fileName)
		}
	},
	Example: `  - bitiful /Users/xxx/xxx.png
  - bitiful "/example/example.png" "/example/example2.png"
  - bitiful -n Test.png /Users/xxx/xxx.png
  - bitiful -n Test.png https://example.com/xxx.png
`,
}

// 创建 S3 会话
func CreateS3Session() {
	SVC = s3.New(session.Must(session.NewSession(
		&aws.Config{
			Region:           aws.String(Region),
			Endpoint:         aws.String(Endpoint),
			S3ForcePathStyle: aws.Bool(false),
			DisableSSL:       aws.Bool(true),
			Credentials: credentials.NewStaticCredentials(
				viper.GetString("AccessKeyID"),
				viper.GetString("AccessKeySecret"),
				"",
			),
		},
	)))
}

// 生成最终上传文件名：有 -n 用 -n，否则用时间戳
func getImgName(image string) {
	fileSuffix := path.Ext(image) // 文件类型
	if fileSuffix == "" {
		fileSuffix = ".png"
	}

	// 有 -n 就用 -n，否则一律生成新名字
	if name != "" {
		fileName = name
		if !strings.HasSuffix(fileName, fileSuffix) {
			fileName += fileSuffix
		}
		return
	}

	fileName = strings.Replace(time.Now().Format("20060102150405.00000"), ".", "", -1) + fileSuffix
}

// 判断文件是否为本地文件
func checkImg(file string) bool {
	if _, err := os.Stat(file); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	} else {
		return true
	}
	return false
}

// 下载图片（URL）
func getImageUrl(imgUrl string) ([]byte, error) {
	resp, err := HttpClient.Get(imgUrl)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("下载图片失败: %v", err))
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Errorf("关闭图片失败: %v", err.Error())
		}
	}(resp.Body)

	// 检查HTTP状态码
	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("下载图片的状态码异常: %v", resp.Status))
	}
	// 读取图片数据
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("读取图片失败: %v", err))
	}
	return data, nil
}

// 读取本地图片文件
func getImageFile(imgPath string) ([]byte, error) {
	tempFile, err := os.Open(imgPath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("打开图片失败: %v", err))
	}
	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			log.Errorf("关闭图片失败: %v", err.Error())
		}
	}(tempFile)
	data, err := io.ReadAll(tempFile)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// 上传图片到 S3
func UploadObject(raw []byte) (output *s3.PutObjectOutput, err error) {
	output, err = SVC.PutObjectWithContext(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(viper.GetString("BucketName")),
		Key:    aws.String(viper.GetString("Path") + fileName),
		Body:   bytes.NewReader(raw),
	})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == request.CanceledErrorCode {
			return nil, errors.New(fmt.Sprintf("上传图片超时: %v", err))
		}
		return nil, errors.New(fmt.Sprintf("上传图片失败: %v", err))
	}
	return output, nil
}

// 程序入口
func main() {
	// 初始化 S3 会话
	CreateS3Session()
	// 注册命令行参数
	rootCmd.Flags().StringVarP(&file, "file", "f", "", "图片 路径|URL .")
	rootCmd.Flags().StringVarP(&name, "name", "n", "", "被覆盖的图片名称.")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "显示详细日志")

	// 日志级别初始化
	cobra.OnInitialize(func() {
		if verbose {
			log.SetLevel(log.DebugLevel)
			log.Debug("已启用详细日志模式")
		} else {
			log.SetLevel(log.InfoLevel)
		}
	})

	if err := rootCmd.Execute(); err != nil {
		log.Errorf("参数错误: %v", err)
	}
}
