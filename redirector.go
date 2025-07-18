package requests

import (
	"errors"
	"net/http"
)

func RedirectPolicyFunc(req *http.Request, via []*http.Request) error {
	if len(via) == 0 {
		return errors.New("via's len should not be 0")
	}
	originReq := via[0]
	// refer: https://stackoverflow.com/questions/16673766/basic-http-auth-in-go
	authorization := originReq.Header.Get("Authorization")
	if authorization != "" {
		req.Header.Add("Authorization", authorization)
	}
	return nil
}
