## Bitiful Typora 图片上传 命令行工具

![show1](https://xrsec.s3.bitiful.net/IMG/2023011808315652353.gif?fmt=webp&q=48)

![show2](https://xrsec.s3.bitiful.net/IMG/2023011808320175950.gif?fmt=webp&q=48)

## 安装与使用

1. [发布版下载](https://github.com/XRSec/bitiful-typora-upload-cli/releases)

```bash
bitiful # 需要配置文件
bitiful "/example/example.png"
bitiful "https://example.com/example.png"
bitiful /path/to/1.png /path/to/2.jpg # 支持多文件批量上传
```

2. 开发构建

```bash
git clone https://github.com/XRSec/bitiful-typora-upload-cli.git
cd bitiful-typora-upload-cli
# 编辑 bitifulUrl 变量
CGO_ENABLED=0 go build -o bitiful
# mv bitiful /usr/local/bin
# cd .. && rm -rf bitiful-typora-upload-cli
```

3. 打开 Typora，`偏好设置` -> `图片` -> `上传图片` -> `自定义命令` -> `/home/xxx/bitiful`

## 配置文件

配置文件路径：`~/Library/Application\ Support/bitiful.yml`

### 配置文件示例

```yaml
Endpoint: s3.bitiful.net
Region: cn-east-1
AccessKeyID: "xxxxxxxxxxx"
AccessKeySecret: "xxxxxxxxxxxxxxxx"
BucketName: "xxxxxxxxxx"
Path: "/xxxxxxxx/"
```

### 指定文件名（覆盖上传）

如需自定义上传后的文件名（如封面图），可用 `-n` 参数：

```bash
bitiful -n "20220903.jpg" https://example.com/20220903.jpg
bitiful -n Test.png /Users/xxx/xxx.png
```

- 仅当指定 `-n` 时，才会用该名字覆盖同名文件。
- 未指定 `-n` 时，自动生成唯一文件名，绝不覆盖。

### 多文件批量上传

直接在命令行追加多个文件或URL即可：

```bash
bitiful /path/to/1.png /path/to/2.jpg
```

### 日志与调试

加 `-v` 参数可显示详细日志，便于排查问题：

```bash
bitiful -v /path/to/1.png
```

## 功能进度

- [x] 本地文件上传
- [x] URL 上传
- [?] ftp/smb/afs
