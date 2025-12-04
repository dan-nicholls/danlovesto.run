package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"net"
	"encoding/json"
	"os/signal"
)

type Config struct {
	ClientID string
	ClientSecret string
	CallbackUrl string
	VerifyToken string
	ListenAddr string
}

type SubscriptionValidation struct {
	Mode string `json:"hub.mode"`
	Challenge string `json:"hub.challenge"`
	VerifyToken string `json:"hub.VerifyToken"`
}

type EventObject struct {
	ObjectType string `json:"object_type"`
	ObjectID string `json:"object_id"`
	AspectType string `json:"aspect_type"`
	Updates string `json:"updates"`
	OwnerID string `json:"owner_id"`
	SubscriptionID int `json:"subscription_id"`
	EventTime int `json:"event_id"` 
}

func decode[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode json: %w", err)
	}
	return v, nil
}

func WebhookHandler(expectedVerifyToken string, sink chan<- EventObject) http.Handler {
	m := map[string]http.Handler{
		http.MethodGet: WebhookVerifyHandler(expectedVerifyToken),
		http.MethodPost: WebhookEventHandler(sink),
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h, ok := m[r.Method]; ok {
			h.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})
}

func WebhookVerifyHandler(expectedVerifyToken string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%v - %v %v\n", r.Host, r.Method, r.URL.Path)
	})
}

func WebhookEventHandler(sink chan<- EventObject) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%v - %v\n", r.Host, r.URL.Path)
		event, err := decode[EventObject](r)
		if err != nil {
			log.Printf("Failed to decode event: %v", err)
			http.Error(w, "Bad Request", http.StatusInternalServerError)	
			return
		}
		sink <- event
	})
}

func PrintEventsWorker(ctx context.Context, in <-chan EventObject) {
	for {
		select {
		case event, ok := <-in:
			if !ok {return}
			log.Printf("Event: %+v\n", event)
		case  <-ctx.Done():
		}
	}
}

func readyHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		log.Printf("%v - %v\n", r.Host, r.URL.Path)
		fmt.Fprintf(w, "hello\n")
	})
}

func StartHTTP(ctx context.Context, listenAddr, verifyToken string, sink chan<- EventObject) (*http.Server, <-chan struct{}, error) {
	// Create Server
	// TODO - Handlers for Health & Ready
	mux := http.NewServeMux()
	mux.Handle("/strava/webhook", WebhookHandler(verifyToken, sink))
	mux.Handle("/ready", readyHandler())

	// TODO - Add logging middleware
	srv := &http.Server {
		Addr: listenAddr,
		ReadTimeout: 5 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout: 60 * time.Second,
		Handler: mux,
	}

	// Check if listen addr is already bound
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Printf("Error binding to listen addr: %v\n", err)
		return nil, nil, err
	}

	ready := make(chan struct{})

	go func() {
		close(ready)

		log.Printf("Starting HTTP Server on %s\n", listenAddr)
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP serve error: %v", err)
		}
	}()
	
	return srv, ready, nil
}


func main() {
	ctx, stop := signal.NotifyContext(context.Background(),os.Interrupt)
	defer stop()

	// Load Config
	cfg := Config{
		ClientID: os.Getenv("CLIENT_ID"),
		ClientSecret : os.Getenv("CLIENT_SECRET"),
		CallbackUrl: os.Getenv("CALLBACK_URI"),
		VerifyToken: os.Getenv("VERIFY_TOKEN"),
		ListenAddr: "0.0.0.0:8080",
	}


	events := make(chan EventObject, 256)

	// Start HTTP Server
	srv, ready, err := StartHTTP(ctx, cfg.ListenAddr, cfg.VerifyToken, events)
	if err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}

	// Start Worker
	go PrintEventsWorker(ctx, events)

	// Check ready
	select {
	case <- ready:
		log.Println("Ready signal recieved from HTTP Server")
	case <- time.After(10 * time.Second):
		log.Fatalf("HTTP Server was not ready in time")
	}

	// Listen for Done signal
	<-ctx.Done()
	log.Println("shutdown signal received")

	// Start Timeout for closing
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	if err := srv.Shutdown(timeoutCtx); err != nil {
		log.Printf("Unable to shutdown server gracefully: %v", err)
	}
	close(events)
	log.Println("Cya!")
}
