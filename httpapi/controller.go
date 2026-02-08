package httpapi

import (
	"encoding/json"
	"github.com/herobeniyoutube/vk-forwarder/application"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type Controller struct {
	handler application.IHandler
}

func NewController(mux *mux.Router, h application.IHandler) Controller {
	c := Controller{h}

	registerMux(mux, &c)

	return c
}

func registerMux(mux *mux.Router, c *Controller) {
	mux.HandleFunc("/", c.helloWorld).Methods("GET")
	mux.HandleFunc("/event", c.verificationHandler).Methods("POST")
}

func (c *Controller) verificationHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err      error
		response *string
	)

	raw, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	var event application.MessageNewEvent
	if err := json.Unmarshal(raw, &event); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	retryCount, _ := strconv.Atoi(r.Header.Get("X-Retry-Counter"))
	ignoreIdempotencyKey := false
	if v := r.Header.Get("Ignore-Idempotency-Key"); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			ignoreIdempotencyKey = b
		}
	}
	response, err = c.handler.Setup(event, retryCount, ignoreIdempotencyKey).Handle()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c.responseOk(w, response)
}

func (c *Controller) helloWorld(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("hello world"))
}

func (c *Controller) responseOk(w http.ResponseWriter, msg *string) {
	r := "ok"
	if msg != nil {
		r = *msg
	}

	log.Printf("response ok: %s", r)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
