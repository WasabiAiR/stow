package oracle

import "time"

type Object struct {
	Bytes        int       `json:"bytes,"`
	ContentType  string    `json:"content-type,"`
	Hash         string    `json:"hash,"`
	LastModified time.Time `json:"last_modified"` // 2015-08-27T09:49:58-05:00 ISO 8601 , CCYY-MM-DDThh:mm:ss±hh:mm
	Name         string    `json:"name,"`
}
