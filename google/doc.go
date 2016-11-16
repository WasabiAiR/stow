// Package google provides stow implementation for Google Cloud Storage.
// In order to use it, you need to make a project in Google, then service account tied
// to that project. Later download a `json` file with the configuration which contents
// you need to pass to the configuration:
//        stow.ConfigMap{
//                "json":       "contents_of_json",
//                "project_id": "google_project_id",
//            }
package google
