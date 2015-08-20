package main

import (
	"io"
	"log"
	"net/http"
	"os"

	"golang.org/x/net/websocket"
)

func main() {
	router()
	http.ListenAndServe(":9000", nil)
}

func router() {
	http.HandleFunc("/", handleIndex)
	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("static"))))
	http.Handle("/message", websocket.Handler(handleMessage))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	fr, err := os.Open("view/index.html")
	if err != nil {
		log.Println("handleIndex:", err)
		http.NotFound(w, r)
		return
	}
	defer fr.Close()

	io.Copy(w, fr)
}
