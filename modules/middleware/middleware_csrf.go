package middleware

import (
	"log"
	"net/http"
	"os"

	gohelper "github.com/fari-99/go-helper"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/csrf"
	adapter "github.com/gwatts/gin-adapter"
	"github.com/spf13/cast"
)

func CsrfMiddleware() gin.HandlerFunc {
	return adapter.Wrap(CsrfServe())
}

func CsrfServe() func(handler http.Handler) http.Handler {
	isSecure := true
	if os.Getenv("APP_STATE") == "test" {
		isSecure = false
	}

	timeExpired := cast.ToInt(os.Getenv("CSRF_EXPIRED"))
	secretKey := gohelper.GenerateRandString(32, "") // random string = one time use token

	csrfMd := csrf.Protect([]byte(secretKey),
		csrf.MaxAge(timeExpired),
		csrf.Secure(isSecure),
		csrf.CookieName("csrf_cookie"),
		csrf.TrustedOrigins([]string{"go-api.fadhlan.loc"}),
		csrf.ErrorHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("%v", r.Header)
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"message": "Forbidden - CSRF token invalid"}`))
		})),
	)

	return csrfMd
}
