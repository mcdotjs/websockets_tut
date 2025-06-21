package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	fmt.Println("web")
	setupApi()
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func setupApi() {
	manager := NewManager()
	http.Handle("/", http.FileServer(http.Dir("./frontend")))
	http.HandleFunc("/ws", manager.ServeWS)
}
