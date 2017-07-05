package s3

import "github.com/aws/aws-sdk-go/service/s3"

type location struct {
	authType    string
	accessKeyID string
	secretKey   string
	region      string
	endpoint    string
	disableSSL  bool
	client      *s3.S3
}
