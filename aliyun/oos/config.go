package oos

import "os"

const (
	endpoint   = "oss-ap-southeast-1.aliyuncs.com"
	BucketName = "bk-sg"
)

var (
	accessID  = os.Getenv("ALIYUN_ACCESS_KEY_ID")
	accessKey = os.Getenv("ALIYUN_ACCESS_KEY_SECRET")
)
