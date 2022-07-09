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
	"github.com/sirupsen/logrus"
)

type apiServer struct {
	ctx  context.Context
	log  *logrus.Entry
	port int
}

func NewAPIServer(ctx context.Context, log *logrus.Entry, port int) *apiServer {
	return &apiServer{
		ctx:  ctx,
		log:  log,
		port: port,
	}
}

func (a *apiServer) Start(backend backend.Backend) error {
	logrus.Infof("Version: %s", version.Get())

	router := mux.NewRouter().StrictSlash(true)
	router.Use(loggingMiddleware(a.log))

	h := newHandler(backend)
	router.Path("/").HandlerFunc(h.root)
	router.Path("/healthz").HandlerFunc(h.root)

	api := router.PathPrefix("/v1").Subrouter()
	api.Path("/domains").Methods("POST").HandlerFunc(h.createDomain)

	authedRoutes := api.PathPrefix("/domains/{domain}").Subrouter()
	authedRoutes.Use(tokenAuthMiddleware(backend))
	authedRoutes.Path("/records").Methods("POST").HandlerFunc(h.createRecord)
	authedRoutes.Path("/records/{record}").Methods("DELETE").HandlerFunc(h.deleteRecord)
	authedRoutes.Path("/renew").Methods("POST").HandlerFunc(h.renew)
	authedRoutes.Methods("GET").HandlerFunc(h.getDomain)
	authedRoutes.Methods("DELETE").HandlerFunc(h.deleteDomain)

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
