package main

import (
	"fmt"
	"net/http"
	"os"
	"context"
	instana "github.com/instana/go-sensor"
	"github.com/opentracing/opentracing-go"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	var span opentracing.Span

	// The HTTP instrumentation injects an entry span into request context, so it can be used
	// as a parent for any operation resulted from an incoming HTTP request. Here we're checking
	// whether the parent span present in request context, which means that our handler has been
	// instrumented.
	if parent, ok := instana.SpanFromContext(r.Context()); ok {
		// Since our handler does some substantial "work", we'd like to have more visibility
		// on how much time it takes to process a request. For this we're starting an _intermediate_
		// span that will be finished as soon as handling is finished.
		span = parent.Tracer().StartSpan("helloHandler", opentracing.ChildOf(parent.Context()))
		defer span.Finish()
	}
	response := os.Getenv("RESPONSE")
	if len(response) == 0 {
		response = "Hello OpenShift!"
	}

	fmt.Fprintln(w, response)
	fmt.Println("Servicing request.")
}

func listenAndServe(port string) {
	fmt.Printf("serving on %s\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}

func main() {
	instana.InitSensor(instana.DefaultOptions())
	
	sensor := instana.NewSensor("golang-ex")
	
	http.HandleFunc("/", instana.TracingHandlerFunc(sensor, "/", helloHandler))
	
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}
	go listenAndServe(port)

	port = os.Getenv("SECOND_PORT")
	if len(port) == 0 {
		port = "8888"
	}
	go listenAndServe(port)

	select {}
}
