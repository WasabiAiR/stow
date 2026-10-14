package s3

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go/logging"
	"github.com/flyteorg/stow"
)

// Kind represents the name of the location/storage type.
const Kind = "s3"

var (
	authTypeAccessKey = "accesskey"
	authTypeIAM       = "iam"
)

const (
	// ConfigAuthType is an optional argument that defines whether to use an IAM role or access key based auth
	ConfigAuthType = "auth_type"

	// ConfigAccessKeyID is one key of a pair of AWS credentials.
	ConfigAccessKeyID = "access_key_id"

	// ConfigSecretKey is one key of a pair of AWS credentials.
	ConfigSecretKey = "secret_key"

	// ConfigToken is an optional argument which is required when providing
	// credentials with temporary access.
	ConfigToken = "token"

	// ConfigRegion represents the region/availability zone of the session.
	ConfigRegion = "region"

	// ConfigEndpoint is optional config value for changing s3 endpoint
	// used for e.g. minio.io
	ConfigEndpoint = "endpoint"

	// ConfigDisableForcePathStyle is optional config value to disable default force path style
	ConfigDisableForcePathStyle = "disable_force_path_style"

	// ConfigDisableSSL is optional config value for disabling SSL support on custom endpoints
	// Its default value is "false", to disable SSL set it to "true".
	ConfigDisableSSL = "disable_ssl"

	// ConfigV2Signing is an optional config value for signing requests with the v2 signature.
	// Its default value is "false", to enable set to "true".
	// This feature is useful for s3-compatible blob stores -- ie minio.
	ConfigV2Signing = "v2_signing"
)

func init() {
	validatefn := func(config stow.Config) error {
		authType, ok := config.Config(ConfigAuthType)
		if !ok || authType == "" {
			authType = authTypeAccessKey
		}

		if !(authType == authTypeAccessKey || authType == authTypeIAM) {
			return errors.New("invalid auth_type")
		}

		if authType == authTypeAccessKey {
			_, ok := config.Config(ConfigAccessKeyID)
			if !ok {
				return errors.New("missing Access Key ID")
			}

			_, ok = config.Config(ConfigSecretKey)
			if !ok {
				return errors.New("missing Secret Key")
			}
		}
		return nil
	}
	makefn := func(config stow.Config) (stow.Location, error) {

		authType, ok := config.Config(ConfigAuthType)
		if !ok || authType == "" {
			authType = authTypeAccessKey
		}

		if !(authType == authTypeAccessKey || authType == authTypeIAM) {
			return nil, errors.New("invalid auth_type")
		}

		if authType == authTypeAccessKey {
			_, ok := config.Config(ConfigAccessKeyID)
			if !ok {
				return nil, errors.New("missing Access Key ID")
			}

			_, ok = config.Config(ConfigSecretKey)
			if !ok {
				return nil, errors.New("missing Secret Key")
			}
		}

		// Create a new client
		client, endpoint, err := newS3Client(config, "")
		if err != nil {
			return nil, err
		}

		// Create a location with given config and client.
		loc := &location{
			config:         config,
			client:         client,
			customEndpoint: endpoint,
		}

		return loc, nil
	}

	kindfn := func(u *url.URL) bool {
		return u.Scheme == Kind
	}

	stow.Register(Kind, makefn, kindfn, validatefn)
}

// Attempts to create a client based on the information given.
func newS3Client(config stow.Config, region string) (client *s3.Client, endpoint string, err error) {
	authType, _ := config.Config(ConfigAuthType)
	accessKeyID, _ := config.Config(ConfigAccessKeyID)
	secretKey, _ := config.Config(ConfigSecretKey)
	token, _ := config.Config(ConfigToken)

	if authType == "" {
		authType = authTypeAccessKey
	}

	if region == "" {
		region, _ = config.Config(ConfigRegion)
	}
	if region == "" {
		region = "us-east-1"
	}

	loadOptions := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithLogger(logging.Nop{}),
		awsconfig.WithRegion(region),
	}
	// The SDK adds the CA bundle of AWS_CA_BUNDLE only to an HTTP client it
	// builds itself, and fails to load the config with any other client.
	if os.Getenv("AWS_CA_BUNDLE") == "" {
		loadOptions = append(loadOptions, awsconfig.WithHTTPClient(http.DefaultClient))
	}
	if authType == authTypeAccessKey {
		loadOptions = append(loadOptions, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKeyID, secretKey, token)))
	}

	awsConfig, err := awsconfig.LoadDefaultConfig(context.Background(), loadOptions...)
	if err != nil {
		return nil, "", err
	}

	disableSSL, _ := config.Config(ConfigDisableSSL)
	disableForcePathStyle, _ := config.Config(ConfigDisableForcePathStyle)
	usev2, _ := config.Config(ConfigV2Signing)
	endpoint, endpointSet := config.Config(ConfigEndpoint)

	s3Client := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		o.EndpointOptions.DisableHTTPS = disableSSL == "true"
		if endpointSet {
			if endpoint != "" {
				o.BaseEndpoint = new(endpointURL(endpoint, disableSSL == "true"))
			}
			o.UsePathStyle = disableForcePathStyle != "true"
			// S3-compatible services do not all accept the checksums the
			// SDK otherwise adds to every request.
			o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
			o.ResponseChecksumValidation = aws.ResponseChecksumValidationWhenRequired
		}
		if usev2 == "true" {
			o.HTTPSignerV4 = v2Signer{}
		}
	})

	return s3Client, endpoint, nil
}

// endpointURL adds the scheme to an endpoint that is configured without one.
func endpointURL(endpoint string, disableSSL bool) string {
	if strings.Contains(endpoint, "://") {
		return endpoint
	}
	if disableSSL {
		return "http://" + endpoint
	}
	return "https://" + endpoint
}
