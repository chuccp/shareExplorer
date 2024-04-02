package web

import (
	auth "github.com/abbot/go-http-auth"
	"github.com/gin-gonic/gin"
	"net/http"
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

//func (digestAuth *DigestAuth) Wrap(wrapped auth.AuthenticatedHandlerFunc) HandlerFunc {
//	handle := digestAuth.DigestAuth.Wrap(wrapped)
//	return func(req *Request) (any, error) {
//		handle.ServeHTTP(req.GetResponseWriter(), req.GetRawRequest())
//		return nil, nil
//	}
//}

func (digestAuth *DigestAuth) Wrap(wrapped auth.AuthenticatedHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if username, authinfo := digestAuth.DigestAuth.CheckAuth(r); username == "" {
			digestAuth.RequireAuth(w, r)
			authenticate := w.Header().Get(digestAuth.Headers.V().Authenticate)
			w.Write([]byte(digestAuth.Headers.V().Authenticate + ":" + authenticate + "\n"))
		} else {
			ar := &auth.AuthenticatedRequest{Request: *r, Username: username}
			if authinfo != nil {
				w.Header().Set(digestAuth.Headers.V().AuthInfo, *authinfo)
			}
			wrapped(w, ar)
		}
	}
}

func (digestAuth *DigestAuth) checkAuth(wrapped auth.AuthenticatedHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, _ := digestAuth.DigestAuth.CheckAuth(r)
		ar := &auth.AuthenticatedRequest{Request: *r, Username: username}
		wrapped(w, ar)
	}
}
func (digestAuth *DigestAuth) CheckAuth(relativePath string, wrapped HandlerFunc) HandlerFunc {
	return func(req *Request) (any, error) {
		var v any
		var err error
		handle := digestAuth.checkAuth(func(writer http.ResponseWriter, request *auth.AuthenticatedRequest) {
			gin.SetMode(gin.ReleaseMode)
			engine := gin.New()
			engine.Any(relativePath, func(context *gin.Context) {
				v, err = wrapped(NewRequest(context, request))
			})
			engine.ServeHTTP(writer, &request.Request)
		})
		handle(req.GetResponseWriter(), req.GetRawRequest())
		return v, err
	}
}
func (digestAuth *DigestAuth) JustCheck(relativePath string, wrapped HandlerFunc) HandlerFunc {
	return func(req *Request) (any, error) {
		var v any
		var err error
		handle := digestAuth.Wrap(func(writer http.ResponseWriter, request *auth.AuthenticatedRequest) {
			gin.SetMode(gin.ReleaseMode)
			engine := gin.New()
			engine.Any(relativePath, func(context *gin.Context) {
				v, err = wrapped(NewRequest(context, request))
			})
			engine.ServeHTTP(writer, &request.Request)
		})
		handle(req.GetResponseWriter(), req.GetRawRequest())
		return v, err
	}
}
func NewDigestAuthenticator(realm string, secrets auth.SecretProvider) *DigestAuth {
	digestAuth := auth.NewDigestAuthenticator(realm, secrets)
	return &DigestAuth{DigestAuth: digestAuth}
}
