package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
)

func main() {
	fmt.Println("web")
	setupApi()
	log.Fatal(http.ListenAndServeTLS(":8080", "server.crt", "server.key", nil))
}

func setupApi() {
	ctx := context.Background()
	manager := NewManager(ctx)

	http.Handle("/", http.FileServer(http.Dir("./frontend")))
	http.HandleFunc("/ws", manager.ServeWS)
	http.HandleFunc("/login", manager.loginHandler)
}
