package requests

type AuthType int

const (
	NoAuth AuthType = iota
	AuthTypeBasic
	AuthTypeProxy  // TODO
	AuthTypeDigest // TODO
)

type AuthInfo struct {
	Type     AuthType
	Username string
	Password string
}
