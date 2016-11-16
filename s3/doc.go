// Package s3 provides stow implementation for Amazon AWS S3 services.
// In order to use it, you need to use a `ConfigMap` with following structure:
//        stow.ConfigMap{
//                "access_key_id": "accessKeyId",
//                "secret_key":    "secretKey",
//                "region":        "region",
//            }
package s3
