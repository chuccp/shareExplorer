package web

import (
	"encoding/base64"
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

func (digestAuth *DigestAuth) Wrap(wrapped auth.AuthenticatedHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if username, authinfo := digestAuth._checkAuth(r); username == "" {
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
func (digestAuth *DigestAuth) filter(r *http.Request) {
	k := r.Header.Get(digestAuth.Headers.V().Authorization)
	if len(k) == 0 {
		auth := r.FormValue("auth")
		decodeString, err := base64.StdEncoding.DecodeString(auth)
		if err == nil {
			r.Header.Set(digestAuth.Headers.V().Authorization, string(decodeString))
		}
	}
}

func (digestAuth *DigestAuth) _checkAuth(r *http.Request) (username string, authinfo *string) {
	digestAuth.filter(r)
	return digestAuth.DigestAuth.CheckAuth(r)
}

func (digestAuth *DigestAuth) ReadAuth(r *http.Request) (username string) {
	digestAuth.filter(r)
	auth1 := auth.DigestAuthParams(r.Header.Get(digestAuth.Headers.V().Authorization))
	return auth1["username"]
}

func (digestAuth *DigestAuth) checkAuth(wrapped auth.AuthenticatedHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, _ := digestAuth._checkAuth(r)
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
			httpMethods := []string{"PROPFIND", "MKCOL", "COPY", "MOVE", "LOCK", "UNLOCK", "PROPPATCH"}
			for _, method := range httpMethods {
				engine.Handle(method, relativePath, func(context *gin.Context) {
					v, err = wrapped(NewRequest(context, request))
				})
			}
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
			httpMethods := []string{"PROPFIND", "MKCOL", "COPY", "MOVE", "LOCK", "UNLOCK", "PROPPATCH"}
			for _, method := range httpMethods {
				engine.Handle(method, relativePath, func(context *gin.Context) {
					v, err = wrapped(NewRequest(context, request))
				})
			}
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
