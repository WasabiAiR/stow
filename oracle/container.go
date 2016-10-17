
package oracle

type ListContainersOutput struct {
	Containers []container
}

type headers struct {
	ContentLength string
	ContentType   string
	Date          string // ISO 8601
	Account       account
	Timestamp     int
	TransactionID string
}

type account struct {
	BytesUsed       int
	ContainerCout   int
	MetaTempURLKey  string
	MetaTempURLKey2 string
	MetaName        string
	ObjectCount     string
}

type container struct {
	Name      string `json:"name,"`
	Count     int    `json:"count,"`
	Bytes     int    `json:"bytes,"`
	AccountID struct {
		ID int `json:"id,"`
	}
	DeleteTimestamp float64 `json:"deleteTimestamp"`
	ContainerID     struct {
		ID int `json:"id,"`
	}
}
