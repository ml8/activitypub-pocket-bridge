package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/acme/autocert"

	"github.com/ml8/ap-bot/activitypub"
	"github.com/ml8/ap-bot/pocket"
)

var (
	port         = flag.String("port", ":8080", "port to listen on")
	certDir      = flag.String("certDir", ".", "directory for certificate storage")
	prod         = flag.Bool("prod", false, "whether to run in production mode")
	domain       = flag.String("domain", "app.jerry.business", "domain for TLS")
	silent       = flag.Bool("silent", false, "whether to silence library logging")
	pocketAppKey = flag.String("pocketAppKey", "", "application key for pocket")
	initUser     = flag.String("initUser", "", "bootstrap user for testing")
	initTok      = flag.String("initTok", "", "bootstrap token for testing")
)

func urlPrefix() string {
	if *prod {
		return "https://" + *domain
	}
	return "http://" + *domain + *port
}

func apUrl() string {
	return urlPrefix() + "/activitypub"
}

func pocketUrl() string {
	return urlPrefix() + "/pocket"
}

func logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		glog.Infof("%s - %s (%s)", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

func main() {
	flag.Parse()

	if *silent {
		glog.Info("Silence system logging, application logging only...")
		log.SetOutput(io.Discard)
	}

	pUrl := pocketUrl()
	b := &pocket.BootstrapData{Users: make(map[string]string)}
	if *initUser != "" {
		b.Users[*initUser] = *initTok
	}
	p := pocket.Init(pUrl, *pocketAppKey, b)

	ap := activitypub.Init(p, activitypub.ResourceMap{
		BaseUrl: apUrl(),
	})

	r := mux.NewRouter()
	r.Use(logger)

	// GET routes
	routes := make(map[string]http.HandlerFunc)
	routes["/pocket"+pocket.RegisterUrlTemplate] = p.RegisterHandler
	routes["/pocket"+pocket.CallbackUrlTemplate] = p.RegisterCallback
	routes["/pocket"+pocket.ArticleUrlTemplate] = p.ArticleHandler
	routes["/activitypub"+activitypub.ActorUrlTemplate] = ap.ActorHandler
	routes[activitypub.WebFingerUrl] = ap.WebFingerHandler

	for u, h := range routes {
		glog.Infof("Registering %v", u)
		r.Handle(u, h)
	}

	// Catch-all 404; log through middleware
	r.NotFoundHandler = r.NewRoute().HandlerFunc(http.NotFound).GetHandler()

	if !*prod {
		glog.Infof("Listening on %v", *port)
		http.ListenAndServe(*port, r)
	} else {
		certManager := autocert.Manager{
			Prompt: autocert.AcceptTOS,
			HostPolicy: func(ctx context.Context, host string) error {
				allowedHost := *domain
				glog.Infof("Host: %v", host)
				if host == allowedHost {
					return nil
				}
				return fmt.Errorf("acme/autocert: only %v is allowed", allowedHost)
			},
			Cache: autocert.DirCache(*certDir),
		}
		server := &http.Server{
			Addr:    ":443",
			Handler: r,
			TLSConfig: &tls.Config{
				GetCertificate: certManager.GetCertificate,
			},
		}
		glog.Infof("Listening...")
		go http.ListenAndServe(":80", certManager.HTTPHandler(nil))
		glog.Fatal(server.ListenAndServeTLS("", ""))
	}
}
