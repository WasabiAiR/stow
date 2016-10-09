package s3

import (
	"bytes"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/cheekybits/is"
	"github.com/graymeta/stow"
	"github.com/graymeta/stow/test"
)

func TestStow(t *testing.T) {
	config := stow.ConfigMap{
		"access_key_id": "AKIAIKXUQN43OZER6ZJQ",
		"secret_key":    "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",
		"region":        "us-west-1",
	}
	test.All(t, "s3", config)
}

func TestEtagCleanup(t *testing.T) {
	etagValue := "9c51403a2255f766891a1382288dece4"
	permutations := []string{
		`"%s"`,       // Enclosing quotations
		`W/\"%s\"`,   // Weak tag identifier with escapted quotes
		`W/"%s"`,     // Weak tag identifier with quotes
		`"\"%s"\"`,   // Double quotes, inner escaped
		`""%s""`,     // Double quotes,
		`"W/"%s""`,   // Double quotes with weak identifier
		`"W/\"%s\""`, // Double quotes with weak identifier, inner escaped
	}
	for index, p := range permutations {
		testStr := fmt.Sprintf(p, etagValue)
		cleanTestStr := cleanEtag(testStr)
		if etagValue != cleanTestStr {
			t.Errorf(`Failure at permutation #%d (%s), result: %s`,
				index, permutations[index], cleanTestStr)
		}
	}
}

// Find some way to PUT an item with metadata, use stow to retrieve it and make assertions
// as needed
func TestMetadata(t *testing.T) {
	is := is.New(t)

	accessKeyID := "AKIAIKXUQN43OZER6ZJQ"
	secretKey := "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"
	region := "us-west-1"
	containerName := "stowtestcontainer" // TODO: randomize
	objectName := "testObject.txt"
	objectContent := []byte("foobarbaz")
	objectLength := int64(len(objectContent))

	awsConfig := aws.NewConfig().
		WithCredentials(credentials.NewStaticCredentials(accessKeyID, secretKey, "")).
		WithRegion(region).
		WithHTTPClient(http.DefaultClient).
		WithMaxRetries(aws.UseServiceDefaultRetries).
		WithLogger(aws.NewDefaultLogger()).
		WithLogLevel(aws.LogOff).
		WithSleepDelay(time.Sleep)

	sess := session.New(awsConfig)
	is.NotNil(sess)

	s3Client := s3.New(sess)
	is.NotNil(s3Client)

	createContainerParams := &s3.CreateBucketInput{
		Bucket: aws.String(containerName),
	}

	_, err := s3Client.CreateBucket(createContainerParams)
	is.NoErr(err)
	defer func() {
		_, err = s3Client.DeleteBucket(&s3.DeleteBucketInput{
			Bucket: &containerName,
		})
		is.NoErr(err)
	}()

	putParams := &s3.PutObjectInput{
		Bucket:        aws.String(containerName),      // Container name
		Key:           (&objectName),                  // Item name
		ContentLength: &objectLength,                  // Item content length
		Body:          bytes.NewReader(objectContent), // Item body
		Metadata: map[string]*string{
			"whos-the-man": aws.String("Bruce Lee"), // metadata as key value
		},
	}

	_, err = s3Client.PutObject(putParams)
	is.NoErr(err)
	defer func() {
		s3Client.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(containerName),
			Key:    aws.String(objectName),
		})
	}()

	// Use Stow to grab the item and its metadata
	config := stow.ConfigMap{
		"access_key_id": "AKIAIKXUQN43OZER6ZJQ",
		"secret_key":    "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",
		"region":        "us-west-1",
	}

	stowLoc, err := stow.Dial("s3", config)
	is.NoErr(err)

	stowCon, err := stowLoc.Container(containerName)
	is.NoErr(err)

	stowIt, err := stowCon.Item(objectName)
	is.NoErr(err)

	itMetadata, err := stowIt.Metadata()
	is.NoErr(err)

	is.Equal(itMetadata["Whos-The-Man"], "Bruce Lee")
}

func TestSetMetadataSuccess(t *testing.T) {
	is := is.New(t)

	m := make(map[string]*string)
	m["one"] = aws.String("two")
	m["3"] = aws.String("4")
	m["ninety-nine"] = aws.String("100")

	mCopy := make(map[string]interface{})
	for key, value := range m {
		mCopy[key] = value
	}

	returnedMap, err := setMetadata(mCopy)
	is.NoErr(err)

	if !reflect.DeepEqual(m, returnedMap) {
		t.Error("Expected and returned maps are not equal.")
	}
}

func TestSetMetadataFailure(t *testing.T) {
	is := is.New(t)

	m := make(map[string]interface{})
	m["name"] = "Corey"
	m["number"] = 9

	_, err := setMetadata(m)
	is.Err(err)
}

func TestItemMetadata(t *testing.T) {
	is := is.New(t)

	containerName := "stowtestintegrationfoobargraymeta"
	objectName := "onetwothree"
	content := "HELLO WORLD"

	objectMetadata := map[string]interface{}{
		"hey": aws.String("you"),
	}

	// Use Stow to grab the item and its metadata
	config := stow.ConfigMap{
		"access_key_id": "AKIAIKXUQN43OZER6ZJQ",
		"secret_key":    "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",
		"region":        "us-west-1",
	}

	stowLoc, err := stow.Dial("s3", config)
	is.NoErr(err)

	stowCon, err := stowLoc.CreateContainer(containerName)
	is.NoErr(err)

	stowCon, err = stowLoc.Container(containerName)
	is.NoErr(err)
	defer func() {
		stowLoc.RemoveContainer(containerName)
	}()

	stowIt, err := stowCon.Put(objectName, strings.NewReader(content), int64(len(content)), objectMetadata)
	is.NoErr(err)

	stowIt, err = stowCon.Item(objectName)
	is.NoErr(err)
	defer func() {
		stowCon.RemoveItem(objectName)
	}()

	itMetadata, err := stowIt.Metadata()
	is.NoErr(err)

	t.Logf("Metadata: %v", itMetadata)
}
