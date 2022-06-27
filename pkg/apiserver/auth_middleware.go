package apiserver

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/acorn-io/acorn-dns/pkg/backend"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type ContextKey string

const DomainID ContextKey = "domainID"

func tokenAuthMiddleware(b backend.Backend) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// createDomain and ping and metrics have no need to check token
			logrus.Debugf("request URL path: %s", r.URL.Path)
			authorization := r.Header.Get("Authorization")
			token := strings.TrimPrefix(authorization, "Bearer ")
			domainName, ok := mux.Vars(r)["domain"]
			if !ok {
				writeError(w, http.StatusForbidden, errors.New("must specify domain"))
			}

			domain, err := b.GetDomain(domainName)
			if err != nil {
				logrus.Errorf("failed to get token hash from DB for %v, err: %v", domainName, err)
				writeError(w, http.StatusForbidden, errors.New("forbidden to use"))
				return
			}

			if domain.TokenHash == "" {
				writeError(w, http.StatusForbidden, errors.New("forbidden to use"))
				return
			}

			if err := bcrypt.CompareHashAndPassword([]byte(domain.TokenHash), []byte(token)); err != nil {
				writeError(w, http.StatusForbidden, errors.New("forbidden to use"))
				return
			}

			ctx := context.WithValue(r.Context(), DomainID, domain.ID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func domainIDFromContext(ctx context.Context) uint {
	domainID, _ := ctx.Value(DomainID).(uint)
	return domainID
}
