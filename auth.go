package requests

type AuthType int

const (
	HTTPNoAuth AuthType = iota
	HTTPBasicAuth
	HTTPProxyAuth  // TODO
	HTTPDigestAuth // TODO
)

type Auth struct {
	authType AuthType
	username string
	password string
}
