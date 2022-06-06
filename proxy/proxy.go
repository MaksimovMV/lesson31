package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	firstInstanceHost  = "http://localhost:8080"
	secondInstanceHost = "http://localhost:8081"
	counter            = 0
)

const addr string = "localhost:8082"

func main() {

	rpURL, err := url.Parse(firstInstanceHost)
	if err != nil {
		log.Fatal(err)
	}
	rp1 := httputil.NewSingleHostReverseProxy(rpURL)
	rpURL, err = url.Parse(secondInstanceHost)
	if err != nil {
		log.Fatal(err)
	}
	rp2 := httputil.NewSingleHostReverseProxy(rpURL)

	server := &http.Server{Addr: addr, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if counter == 0 {
			rp1.ServeHTTP(w, r)
			counter++
		} else {
			rp2.ServeHTTP(w, r)
			counter--
		}
	})}
	// Server run context
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, _ := context.WithTimeout(serverCtx, 30*time.Second)

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()

		// Trigger graceful shutdown
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}
		serverStopCtx()
	}()

	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
}
