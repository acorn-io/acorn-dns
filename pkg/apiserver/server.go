package apiserver

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/acorn-io/acorn-dns/pkg/backend"
	"github.com/acorn-io/acorn-dns/pkg/version"
	ghandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/sethvargo/go-limiter/httplimit"
	"github.com/sethvargo/go-limiter/memorystore"
	"github.com/sirupsen/logrus"
)

type apiServer struct {
	ctx               context.Context
	log               *logrus.Entry
	port              int
	unauthedRateLimit uint64
	authedRateLimit   uint64
	rateLimitState    string
}

func NewAPIServer(ctx context.Context, log *logrus.Entry, port int, unauthedRateLimit, authedRateLimt uint64, rateLimitMode string) *apiServer {
	return &apiServer{
		ctx:               ctx,
		log:               log,
		port:              port,
		unauthedRateLimit: unauthedRateLimit,
		authedRateLimit:   authedRateLimt,
		rateLimitState:    rateLimitMode,
	}
}

func (a *apiServer) Start(backend backend.Backend) error {
	logrus.Infof("Version: %s", version.Get())

	router := mux.NewRouter().StrictSlash(true)
	router.Use(loggingMiddleware(a.log))
	h := newHandler(backend)

	// When functioning properly, these routes will return the version of tha app that is running
	router.Path("/").HandlerFunc(h.root)
	router.Path("/healthz").HandlerFunc(h.root)

	api := router.PathPrefix("/v1").Subrouter()

	// POSTing to a domain creates a new domain and token. Further requests (below) against the created domain resource
	// require authentication using the token. This gets rate limited by IP to prevent spamming creation of domains
	unauthedLimiter := newRateLimiter(a.unauthedRateLimit, a.rateLimitState, httplimit.IPKeyFunc("X-Forwarded-For", "X-Real-IP"))
	api.Path("/domains").Methods("POST").Handler(unauthedLimiter.wrap(h.createDomain))

	// All routes using this authedRoutes subrouter will require token based authentication
	authedRoutes := api.PathPrefix("/domains/{domain}").Subrouter()
	authedRoutes.Use(tokenAuthMiddleware(backend))

	// Basic routes for the domain resource
	authedRoutes.Methods("GET").HandlerFunc(h.getDomain)

	authedLimiter := newRateLimiter(a.authedRateLimit, a.rateLimitState, domainKeyFunc)

	// These are for records sub-resource
	authedRoutes.Path("/records").Methods("POST").Handler(authedLimiter.wrap(h.createRecord))
	authedRoutes.Path("/records/{record}").Methods("DELETE").Handler(authedLimiter.wrap(h.deleteRecord))

	// These are "actions" that can be taken on a domain
	authedRoutes.Path("/renew").Methods("POST").Handler(authedLimiter.wrap(h.renew))
	authedRoutes.Path("/purgerecords").Methods("POST").Handler(authedLimiter.wrap(h.purgerecords))

	// Note: this allows not found urls to be logged via the middleware
	// It **HAS** to be defined after all other paths are defined.
	router.NotFoundHandler = router.NewRoute().HandlerFunc(http.NotFound).GetHandler()

	// Below this point is where the server is started and graceful shutdown occurs.
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", a.port),
		Handler: ghandlers.CORS()(router),
	}

	go func() {
		a.log.WithField("port", a.port).Info("starting api server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.log.Fatalf("listen: %s\n", err)
		}
	}()

	go backend.StartPurgerDaemon(a.ctx.Done())

	<-a.ctx.Done()

	a.log.Info("shutting down the api server gracefully")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	if err := srv.Shutdown(ctx); err != nil {
		a.log.WithError(err).Error("unable to shutdown the api server gracefully")
		return err
	}

	return nil
}

func newRateLimiter(tokens uint64, state string, keyFunc func(r *http.Request) (string, error)) *rl {
	store, err := memorystore.New(&memorystore.Config{
		Tokens:   tokens,
		Interval: time.Hour,
	})

	mw, err := httplimit.NewMiddleware(store, keyFunc)
	if err != nil {
		logrus.Fatal("middleware store or keyFunc nil")
	}

	rl := &rl{
		state: state,
		mw:    mw,
	}

	return rl
}

type rl struct {
	mw    *httplimit.Middleware
	state string
}

func (r *rl) wrap(next func(w http.ResponseWriter, r *http.Request)) http.Handler {
	if r.state == "enabled" {
		return r.mw.Handle(http.HandlerFunc(next))
	} else {
		logrus.Warnf("Rate limiting disabled")
	}
	return http.HandlerFunc(next)
}

func domainKeyFunc(r *http.Request) (string, error) {
	vars := mux.Vars(r)
	domainName := vars["domain"]
	return domainName, nil
}
