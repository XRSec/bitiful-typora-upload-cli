package main

import (
	"bytes"
	"context"
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
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

var (
	Endpoint    = "s3.ladydaily.com"
	SVC         *s3.S3
	update      bool
	file        string
	name        string
	HttpClient  *http.Client
	versionData string
	fileName    string
	imageType   bool
	err         error
)

var rootCmd = &cobra.Command{
	Short:   "TYPORA UPLOAD IAMGE TO AWS",
	Long:    "优雅的使用 TYPORA 上传图片到 AWS 存储桶",
	Version: versionData,

	Run: func(cmd *cobra.Command, args []string) {
		for _, v := range args {
			var data []byte
			getImgName(v)
			imageType = checkImg(v)
			// 判断文件是否存在
			if !imageType {
				data, err = getImageUrl(v)
				if err != nil {
					log.Fatalln(err)
					continue
				}
			} else {
				data, err = getImageFile(v)
				if err != nil {
					log.Fatalln(err)
					continue
				}
			}

			_, err := UploadObject(data)
			if err != nil {
				log.Fatalln(err)
				continue
			}

			if imageType {
				if err := wastebasket.Trash(v); err != nil {
					log.Errorf("删除文件失败: %v", err)
					return
				}
			}
			log.Infof("Success: https://%v.%v%v%v?fmt=webp&q=48", viper.GetString("BucketName"), Endpoint, viper.GetString("Path"), fileName)

		}
	},
	Example: `  - bitiful /Users/xxx/xxx.png
  - bitiful "/example/example.png" "/example/example2.png"
  - bitiful -u -f/Users/xxx/xxx.png -nTest.png
  - bitiful -u -f /Users/xxx/xxx.png -n Test.png
`,
}

func CreateS3Session() {
	SVC = s3.New(session.Must(session.NewSession(
		&aws.Config{
			Region:           aws.String("cn-north-1"),
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

func getImgName(image string) {
	//filetype := http.DetectContentType(data)
	var imgType string
	fileNameAll := path.Base(image)
	fileSuffix := path.Ext(image)
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
			log.Errorf("不支持的图片格式: %v", fileSuffix)
		}
	}

	// 检测是否覆盖
	if update {
		fileName = name + imgType
	} else {
		fileName = strings.Replace(time.Now().Format("20060102150405.00000"), ".", "", -1) + imgType
	}
}

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

	// Check Status
	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("下载图片的状态码异常: %v", resp.Status))
	}
	// Read the data
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("读取图片失败: %v", err))
	}
	return data, nil
}

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

func UploadObject(raw []byte) (output *s3.PutObjectOutput, err error) {
	// Upload to s3
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

func main() {
	// upload img/png/jpeg/webp to aws
	CreateS3Session()
	rootCmd.Flags().BoolVarP(&update, "update", "u", false, "更新您的图片.")
	rootCmd.Flags().StringVarP(&file, "file", "f", "", "图片 路径|URL .")
	rootCmd.Flags().StringVarP(&name, "name", "n", "", "被覆盖的图片名称.")
	if err := rootCmd.Execute(); err != nil {
		log.Errorf("参数错误: %v", err)
	}
}