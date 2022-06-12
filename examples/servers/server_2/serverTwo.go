package main

import (
	"fmt"
	"net/http"
	"os"
)

const (
	PORT = 2002
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Server 2")
	fmt.Fprintf(w, "Server 2")
}

func main() {
	serverName := os.Getenv("SERVER_NAME")
	hostAndPort := fmt.Sprintf("localhost:%d", PORT)
	http.HandleFunc("/", handler)
	fmt.Printf("Server running on http://%s\n", hostAndPort)
	fmt.Printf("Server name:%s\n", serverName)
	http.ListenAndServe(hostAndPort, nil)
}
