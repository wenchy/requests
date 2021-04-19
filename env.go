package requests

type Env struct {
	Timeout int64
}

var env Env

func init() {
	env.Timeout = 60 // seconds
}

func SetEnvTimeout(timeout int64) {
	env.Timeout = timeout
}
