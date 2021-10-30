package support

import (
	"github.com/gorilla/securecookie"
	"github.com/iris-contrib/middleware/cors"
	"github.com/kataras/golog"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/sessions"
	"strings"
	"time"
)

var (
	CookieName      = "PLAYXY-NATOK"
	SessionKey      = "PLAYXY-NATOK-AUTHENTICATION"
	SessionsManager *sessions.Sessions
)

// TimeCounter 时间内计数器
type TimeCounter struct {
	time    time.Time
	counter int64
}

func init() {
	//附加session管理器
	// AES仅支持16,24或32字节的密钥大小。
	// 您需要准确提供该字节数，或者从您键入的内容中获取密钥。
	hashKey := []byte("the-big-and-secret-fash-key-here")
	blockKey := []byte("lot-secret-of-characters-big-too")
	secureCookie := securecookie.New(hashKey, blockKey)
	SessionsManager = sessions.New(sessions.Config{
		Cookie:  CookieName,
		Encode:  secureCookie.Encode,
		Decode:  secureCookie.Decode,
		Expires: 24 * time.Hour,
	})
}

var (
	basics  = []string{"/static", "/js", "/css", "/favicon.ico", "/captcha"}
	passUri = []string{"/user/login", "/index.html"}
)

// AuthorHandler 认证拦截器
func AuthorHandler() func(iris.Context) {
	return func(ctx iris.Context) {
		path := ctx.Path()
		for _, pass := range append(basics, passUri...) {
			if strings.HasPrefix(path, pass) {
				ctx.Next()
				return
			}
		}
		if nil != SessionsManager.Start(ctx).Get(SessionKey) {
			ctx.Next()
			return
		}
		ctx.Next()
		golog.Warn("Warn intercept: ", path)
		//ctx.Redirect(loginWeb, iris.StatusFound)
	}
}

// CorsHandler 开启可跨域
func CorsHandler() func(iris.Context) {
	return cors.New(cors.Options{
		AllowCredentials: true,
		AllowedOrigins:   []string{"*"},
		AllowedHeaders:   []string{"*"},
		AllowedMethods:   []string{"HEAD", "GET", "POST", "PUT", "PATCH", "DELETE"},
	})
}
