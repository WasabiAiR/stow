/*
Copyright (c) 2013 Damien Le Berrigaud and Nick Wade

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package s3

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
)

var s3ParamsToSign = map[string]bool{
	"acl":                          true,
	"location":                     true,
	"logging":                      true,
	"notification":                 true,
	"partNumber":                   true,
	"policy":                       true,
	"tagging":                      true,
	"requestPayment":               true,
	"torrent":                      true,
	"uploadId":                     true,
	"uploads":                      true,
	"versionId":                    true,
	"versioning":                   true,
	"versions":                     true,
	"response-content-type":        true,
	"response-content-language":    true,
	"response-expires":             true,
	"response-cache-control":       true,
	"response-content-disposition": true,
	"response-content-encoding":    true,
	"website":                      true,
	"delete":                       true,
}

// v2Signer signs requests with signature version 2. The SDK only knows
// signature version 4, so it is plugged in as the client's version 4 signer.
type v2Signer struct{}

// SignHTTP signs the request with an Authorization header.
func (v2Signer) SignHTTP(_ context.Context, credentials aws.Credentials, r *http.Request, _ string, _ string, _ string,
	signingTime time.Time, _ ...func(*v4.SignerOptions)) error {
	r.Header["Host"] = []string{r.URL.Host}
	r.Header["x-amz-date"] = []string{signingTime.In(time.UTC).Format(time.RFC1123)}

	signature := signV2(credentials, r, "")
	r.Header["Authorization"] = []string{"AWS " + credentials.AccessKeyID + ":" + signature}
	return nil
}

// PresignHTTP signs the request with query parameters. The SDK passes the
// lifetime of the URL in the version 4 X-Amz-Expires parameter.
func (v2Signer) PresignHTTP(_ context.Context, credentials aws.Credentials, r *http.Request, _ string, _ string, _ string,
	signingTime time.Time, _ ...func(*v4.SignerOptions)) (string, http.Header, error) {
	params := r.URL.Query()
	lifetime, err := strconv.ParseInt(params.Get("X-Amz-Expires"), 10, 64)
	if err != nil {
		return "", nil, err
	}
	expires := strconv.FormatInt(signingTime.Unix()+lifetime, 10)

	params.Del("X-Amz-Expires")
	params.Set("AWSAccessKeyId", credentials.AccessKeyID)
	params.Set("Expires", expires)
	r.URL.RawQuery = params.Encode()

	params.Set("Signature", signV2(credentials, r, expires))
	r.URL.RawQuery = params.Encode()
	return r.URL.String(), r.Header, nil
}

// signV2 returns the version 2 signature of the request. A presigned request
// is signed with its expiry time instead of its date.
func signV2(credentials aws.Credentials, r *http.Request, expires string) string {
	var (
		md5, ctype, xamz string
		sarray           []string
	)
	date := expires

	for k, v := range r.Header {
		k = strings.ToLower(k)
		switch k {
		case "content-md5":
			md5 = v[0]
		case "content-type":
			ctype = v[0]
		default:
			if strings.HasPrefix(k, "x-amz-") {
				sarray = append(sarray, k+":"+strings.Join(v, ","))
			}
		}
	}
	if len(sarray) > 0 {
		sort.Strings(sarray)
		xamz = strings.Join(sarray, "\n") + "\n"
	}

	canonicalPath := r.URL.EscapedPath()
	sarray = sarray[0:0]
	for k, v := range r.URL.Query() {
		if s3ParamsToSign[k] {
			for _, vi := range v {
				if vi == "" {
					sarray = append(sarray, k)
				} else {
					sarray = append(sarray, k+"="+vi)
				}
			}
		}
	}
	if len(sarray) > 0 {
		sort.Strings(sarray)
		canonicalPath = canonicalPath + "?" + strings.Join(sarray, "&")
	}

	stringToSign := strings.Join([]string{
		r.Method,
		md5,
		ctype,
		date,
		xamz + canonicalPath,
	}, "\n")
	hash := hmac.New(sha1.New, []byte(credentials.SecretAccessKey))
	hash.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}
