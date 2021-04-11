package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {

	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	http.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("bar"))
	})

	host := os.Getenv("MYAPP_HOST")
	fmt.Printf("Host: %s\n", host)

	port := os.Getenv("MYAPP_PORT")
	fmt.Printf("Port: %s\n", port)

	addr := fmt.Sprintf("%s:%s", host, port)
	fmt.Printf("Server address: %s\n", addr)

	http.ListenAndServe(addr, nil)
}
