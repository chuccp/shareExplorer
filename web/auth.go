package web

import (
	auth "github.com/abbot/go-http-auth"
)

type BasicAuth struct {
	*auth.BasicAuth
}

func (basicAuth *BasicAuth) Wrap(wrapped auth.AuthenticatedHandlerFunc) HandlerFunc {
	handle := basicAuth.BasicAuth.Wrap(wrapped)
	return func(req *Request) (any, error) {
		handle.ServeHTTP(req.GetResponseWriter(), req.GetRawRequest())
		return nil, nil
	}
}

type DigestAuth struct {
	*auth.DigestAuth
}

func (digestAuth *DigestAuth) Wrap(wrapped auth.AuthenticatedHandlerFunc) HandlerFunc {
	handle := digestAuth.DigestAuth.Wrap(wrapped)
	return func(req *Request) (any, error) {
		handle.ServeHTTP(req.GetResponseWriter(), req.GetRawRequest())
		return nil, nil
	}
}

func NewBasicAuthenticator(realm string, secrets auth.SecretProvider) *BasicAuth {
	basicAuth := auth.NewBasicAuthenticator(realm, secrets)
	return &BasicAuth{BasicAuth: basicAuth}
}
func NewDigestAuthenticator(realm string, secrets auth.SecretProvider) *DigestAuth {
	digestAuth := auth.NewDigestAuthenticator(realm, secrets)
	return &DigestAuth{DigestAuth: digestAuth}
}
