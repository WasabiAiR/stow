package s3

import (
	"net/http"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/graymeta/stow"
	"github.com/pkg/errors"
)

// Kind represents the name of the location/storage type.
const Kind = "s3"

const (
	// ConfigAccessKeyID is one key of a pair of AWS credentials.
	ConfigAccessKeyID = "access_key_id"

	// ConfigSecretKey is one key of a pair of AWS credentials.
	ConfigSecretKey = "secret_key"

	// ConfigToken is an optional argument which is required when providing
	// credentials with temporary access.
	// ConfigToken = "token"

	// ConfigRegion represents the region/availability zone of the session.
	ConfigRegion = "region"
)

func init() {

	makefn := func(config stow.Config) (stow.Location, error) {
		_, ok := config.Config(ConfigAccessKeyID)
		if !ok {
			return nil, errors.New("missing Access Key ID")
		}

		_, ok = config.Config(ConfigSecretKey)
		if !ok {
			return nil, errors.New("missing Secret Key")
		}

		_, ok = config.Config(ConfigRegion)
		if !ok {
			return nil, errors.New("missing Region")
		}

		// Create a new client (s3 session)
		client, err := newS3Client(config)
		if err != nil {
			return nil, err
		}

		// Create a location with given config and client (s3 session).
		loc := &location{
			config: config,
			client: client,
		}

		return loc, nil
	}

	kindfn := func(u *url.URL) bool {
		return u.Scheme == Kind
	}

	stow.Register(Kind, makefn, kindfn)
}

// Attempts to create a session based on the information given.
func newS3Client(config stow.Config) (*s3.S3, error) {
	accessKeyID, _ := config.Config(ConfigAccessKeyID)
	secretKey, _ := config.Config(ConfigSecretKey)
	//	token, _ := config.Config(ConfigToken)
	region, _ := config.Config(ConfigRegion)

	awsConfig := aws.NewConfig().
		WithCredentials(credentials.NewStaticCredentials(accessKeyID, secretKey, "")).
		WithRegion(region).
		WithHTTPClient(http.DefaultClient).
		WithMaxRetries(aws.UseServiceDefaultRetries).
		WithLogger(aws.NewDefaultLogger()).
		WithLogLevel(aws.LogOff).
		WithSleepDelay(time.Sleep)

	sess := session.New(awsConfig)
	if sess == nil {
		return nil, errors.New("creating the S3 session")
	}

	s3Client := s3.New(sess)

	return s3Client, nil
}
