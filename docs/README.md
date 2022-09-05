## Bitful Typora Upload Tools

## Install & Use

1. [Release Download](https://github.com/XRSec/bitful-typora-upload-tools/releases)

```bash
bitful-typora-upload-tools -u "BitfulUrl"
```

2. dev build

```bash
git clone https://github.com/XRSec/bitful-typora-upload-tools.git
cd bitful-typora-upload-tools
# Edit BitfulUrl in bitful-typora-upload-tools.go
CGO_ENABLED=0 go build -o bitful-typora-upload-tools
# mv bitful-typora-upload-tools /usr/local/bin
# cd .. && rm -rf bitful-typora-upload-tools
```

3. Open Typora and set `Tools` -> `Preference` -> `Image` -> `Upload Image` -> `Custom Command` -> `/home/xxx/bitful-typora-upload-tools`

### Generating configuration files

```bash
bitful-typora-upload-tools -u "BitfulUrl"
```

### Cover image

```bash
bitful-typora-upload-tools -c "20220903" https://example.com/20220903.jpg
```
### High-definition pictures

```bash
bitful-typora-upload-tools.go > UploadFile > fmt.Println(result.Data.Url) ==> fmt.Println(result.Data.OUrl)
```

or delete `?w=1280&fmt=jpg`

## Doing

- [x] file
- [x] url
- [?] ftp/smb/afs

## Todo
- [ ] cli
