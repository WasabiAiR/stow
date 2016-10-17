package oracle

const Kind = "oracle"

const (
	// <service type>-<namespace>:<username>
	// storage-a422618:corey@graymeta.com
	storageUsername = "username"

	// Raw text
	// foobar
	storagePassword = "password"

	// Obtained from container information page. Note, must be modified.
	// https://storage-a422618.storage.oraclecloud.com/auth/v1.0
	authEndpoint = "endpoint"
)

