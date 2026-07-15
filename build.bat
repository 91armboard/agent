set GO111MODULE=on
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=mipsle
set GOMIPSLE=softfloat
set GOMIPS=softfloat
go build -ldflags "-s -w" -o ss_main.upgrade main.go
copy ss_main.upgrade main
CertUtil -hashfile main MD5 | findstr /i /v "MD5 hash of" > md5.upgrade 