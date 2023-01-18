## Bitiful Typora Upload CLI

![show1](https://xrsec.s3.ladydaily.com/IMG/2023011808315652353.gif?fmt=webp&q=48)

![show2](https://xrsec.s3.ladydaily.com/IMG/2023011808320175950.gif?fmt=webp&q=48)

## Install & Use

1. [Release Download](https://github.com/XRSec/bitiful-typora-upload-cli/releases)

```bash
bitiful # require config file
bitiful "/example/example.png"
bitiful "https://example.com/example.png"
```

or dev build

```bash
git clone https://github.com/XRSec/bitiful-typora-upload-cli.git
cd bitiful-typora-upload-cli
# Edit bitifulUrl in bitiful-typora-upload-cli.go
CGO_ENABLED=0 go build -o bitiful
# mv bitiful /usr/local/bin
# cd .. && rm -rf bitiful-typora-upload-cli
```

2. Open Typora and set `cli` -> `Preference` -> `Image` -> `Upload Image` -> `Custom Command` -> `/home/xxx/bitiful`

## Config

```yaml
AccessKeyID: "xxxxxxxxxxx"
AccessKeySecret: "xxxxxxxxxxxxxxxx"
BucketName: "xxxxxxxxxx"
Path: "/xxxxxxxx/"
```

### Cover image

```bash
bitiful -u -n "20220903" -f https://example.com/20220903.jpg
bitiful -u -f/Users/xxx/xxx.png -nTest.png
```
## Doing

- [x] file
- [x] url
- [?] ftp/smb/afs
