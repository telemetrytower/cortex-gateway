package gateway

import (
	"errors"
	"fmt"
	"net/http"

	util_log "github.com/cortexproject/cortex/pkg/util/log"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	jwt "github.com/golang-jwt/jwt/v4"
	jwtReq "github.com/golang-jwt/jwt/v4/request"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/weaveworks/common/middleware"

	"github.com/telemetrytower/cortex-gateway/pkg/org"
)

var (
	authFailures = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "cortex_gateway",
		Name:      "failed_authentications_total",
		Help:      "The total number of failed authentications.",
	}, []string{"reason"})
	authSuccess = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "cortex_gateway",
		Name:      "succeeded_authentications_total",
		Help:      "The total number of succeeded authentications.",
	}, []string{"tenant"})
)

func NewAuthenticate(jwtSecret string, basics map[string]BasicAuth) middleware.Func {
	enableBasicAuth := len(basics) > 0

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := log.With(util_log.WithContext(r.Context(), util_log.Logger), "ip_address", r.RemoteAddr)
			level.Debug(logger).Log("msg", "authenticating request", "route", r.RequestURI)

			tokenString := r.Header.Get("Authorization") // Get operation is case insensitive
			if tokenString == "" {
				level.Info(logger).Log("msg", "no Authorization token provided")
				http.Error(w, "No Authorization token provided", http.StatusUnauthorized)
				authFailures.WithLabelValues("no_token").Inc()
				return
			}

			var (
				tenantID string
				err      error
			)

			if enableBasicAuth {
				tenantID, err = basicAuth(w, r, logger, basics)
			}
			if err != nil {
				return
			}

			// Try Jwt auth
			if tenantID == "" {
				tenantID, err = jwtAuth(w, r, logger, jwtSecret)
			}
			if err != nil {
				return
			}

			if tenantID == "" {
				level.Error(logger).Log("msg", "empty tenant resolve from Authorization")
				http.Error(w, "Invalid Authorization token provided", http.StatusUnauthorized)
				return
			}

			// Token is valid
			authSuccess.WithLabelValues(tenantID).Inc()
			r.Header.Set("X-Scope-OrgID", tenantID)
			next.ServeHTTP(w, r)
		})
	}
}

func basicAuth(w http.ResponseWriter, r *http.Request, logger log.Logger, basics map[string]BasicAuth) (string, error) {
	username, password, ok := r.BasicAuth()
	if !ok {
		return "", nil
	}

	expectAuth, exist := basics[username]
	if exist && expectAuth.Password == password {
		return expectAuth.Tenant, nil
	}

	return "", errors.New("invalid Basic Authorization")
}

func jwtAuth(w http.ResponseWriter, r *http.Request, logger log.Logger, jwtSecret string) (string, error) {
	// Try to parse and validate JWT
	te := &org.Tenant{}
	_, err := jwtReq.ParseFromRequest(
		r,
		jwtReq.AuthorizationHeaderExtractor,
		func(token *jwt.Token) (interface{}, error) {
			// Only HMAC algorithms accepted - algorithm validation is super important!
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				level.Info(logger).Log("msg", "unexpected signing method", "used_method", token.Header["alg"])
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}

			return []byte(jwtSecret), nil
		},
		jwtReq.WithClaims(te))

	// If Tenant's Valid method returns false an error will be set as well, hence there is no need
	// to additionally check the parsed token for "Valid"
	if err != nil {
		level.Info(logger).Log("msg", "invalid bearer token", "err", err.Error())
		http.Error(w, "Invalid bearer token", http.StatusUnauthorized)
		authFailures.WithLabelValues("token_not_valid").Inc()
	}

	return te.TenantID, err
}
