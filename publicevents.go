package main

import (
	"log"
	"net/http"
)

type publicEvent struct {
	Event  string
	Params map[string]interface{}
}

func publicEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		log.Println("internal error: ResponseWriter not a http.Flusher")
		http.Error(w, "internal error", 500)
		return
	}

	closenotifier, ok := w.(http.CloseNotifier)
	if !ok {
		log.Println("internal error: ResponseWriter not a http.CloseNotifier")
		http.Error(w, "internal error", 500)
		return
	}
	closenotify := closenotifier.CloseNotify()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	_, err := w.Write([]byte(": keepalive\n\n"))
	if err != nil {
		return
	}
	flusher.Flush()

	events := make(chan publicEvent, 8)

	defer func() {
		ctl.unregisterPublicEvents <- events
	}()

	ctl.registerPublicEvents <- events
	for {
		select {
		case e, ok := <-events:
			if !ok {
				writePublicEvent(w, publicEvent{Event: "closed"})
				flusher.Flush()
				return
			}
			err = writePublicEvent(w, e)
			flusher.Flush()
			if err != nil {
				log.Printf("writing event: %s\n", err)
				return
			}
		case <-closenotify:
			return
		}
	}
}
