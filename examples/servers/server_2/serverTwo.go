package main

import (
	"fmt"
	"net/http"
)

const (
	PORT = 2002
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Server 2")
}

func main() {
	hostAndPort := fmt.Sprintf("localhost:%d", PORT)
	http.HandleFunc("/", handler)
	fmt.Printf("Server running on http://%s\n", hostAndPort)
	http.ListenAndServe(hostAndPort, nil)
}
