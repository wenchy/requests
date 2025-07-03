package auth

type Type int

const (
	NoAuth Type = iota
	BasicAuth
	ProxyAuth  // TODO
	DigestAuth // TODO
)

type AuthInfo struct {
	Type     Type
	Username string
	Password string
}
