package s3

import (
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/graymeta/stow"
	"github.com/pkg/errors"
)

const (
	// Kind represents the name of the storage type.
	Kind = "s3"

	// ConfigRegion represents the region/availability zone of the session.
	// Defaults to "us-east-1".
	ConfigRegion = "region"

	// ConfigEndpoint is an optional config value for changing s3 endpoint
	// used with e.g. minio.io.
	ConfigEndpoint = "endpoint"

	// ConfigAccessKeyID is one key of a pair of AWS credentials.
	ConfigAccessKeyID = "access_key_id"

	// ConfigSecretKey is one key of a pair of AWS credentials.
	ConfigSecretKey = "secret_key"

	// ConfigAuthType is an optional argument that defines whether to use an IAM role or access key based auth
	ConfigAuthType = "auth_type"

	// ConfigDisableSSL is optional config value for disabling SSL support on custom endpoints
	// Its default value is "false", to disable SSL set it to "true".
	ConfigDisableSSL = "disable_ssl"
)

func init() {

	makefn := func(config stow.Config) (stow.Location, error) {

		awsConfig := aws.NewConfig()
		awsConfig.WithLogger(aws.NewDefaultLogger()).WithLogLevel(aws.LogOff)

		authType, ok := config.Config(ConfigAuthType)
		if !ok || authType == "" {
			authType = "accesskey"
		}

		if authType != "accesskey" && authType != "iam" {
			return nil, errors.New("invalid auth_type")
		}
		var accessKeyID string
		var secretKey string
		if authType == "accesskey" {
			var ok bool
			if accessKeyID, ok = config.Config(ConfigAccessKeyID); !ok {
				return nil, errors.New("missing access_key_id")
			}

			if secretKey, ok = config.Config(ConfigSecretKey); !ok {
				return nil, errors.New("missing secret_key")
			}
			awsConfig.WithCredentials(credentials.NewStaticCredentials(accessKeyID, secretKey, ""))
		}

		region, ok := config.Config(ConfigRegion)
		if !ok {
			region = "us-east-1"
		}
		awsConfig.WithRegion(region)

		endpoint, ok := config.Config(ConfigEndpoint)
		if ok {
			awsConfig.WithEndpoint(endpoint).WithS3ForcePathStyle(true)
		}

		var disSSL bool
		disableSSL, ok := config.Config(ConfigDisableSSL)
		if ok && disableSSL == "true" {
			disSSL = true
		}
		awsConfig.WithDisableSSL(disSSL)

		// todo(piotr): make authtype its own enum-like type

		sess, err := session.NewSession(awsConfig)
		if err != nil {
			return nil, errors.Wrap(err, "creating S3 session")
		}

		c := s3.New(sess)

		loc := &location{
			authType:    authType,
			accessKeyID: accessKeyID,
			secretKey:   secretKey,
			region:      region,
			endpoint:    endpoint,
			disableSSL:  disSSL,
			client:      c,
		}

		return loc, nil
	}

	kindfn := func(u *url.URL) bool {
		return u.Scheme == Kind
	}

	stow.Register(Kind, makefn, kindfn)
}
