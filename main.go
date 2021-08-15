package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func main() {

	ch1 := make(chan string)
	var schedule func()
	schedule = func() {
		time.AfterFunc(10*time.Second, func() {

			t := time.Now()
			log.Printf("scheduled runnign %s", t)
			ch1 <- t.Format("2006-01-02")
			schedule()
		})
	}

	schedule()

	root := http.NewServeMux()
	root.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		page, err := ioutil.ReadFile("index.html")
		if err != nil {
			log.Fatal(err)
		}

		w.Header().Set("Content-Length", fmt.Sprint(len(page)))
		fmt.Fprintf(w, "%s", page)

	})

	root.Handle("/event", serverSentEvent(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			log.Fatal("does not support flushing")
		}

		for {

			select {
			case event := <-ch1:
				log.Println("received", event)

				fmt.Fprint(w, writeEvent("message", event))
				flusher.Flush()
			case <-r.Context().Done():
				log.Println("goodbye")
				return

			}

		}

	})))

	log.Println("listening on :8080")
	if err := http.ListenAndServe(":8080", root); err != nil {
		log.Fatal(err)
	}
}

func writeEvent(event, data string) string {
	return fmt.Sprintf("event: %s\ndata: %s\n\n", event, data)

}

func serverSentEvent(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
