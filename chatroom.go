package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"golang.org/x/net/websocket"
)

var (
	gPort   int
	gIsHelp bool
)

func init() {
	flag.IntVar(&gPort, "p", 10000, "web server port")
	flag.IntVar(&gPort, "port", 10000, "web server port")
	flag.BoolVar(&gIsHelp, "h", false, "show help")
	flag.BoolVar(&gIsHelp, "help", false, "show help")
}

func main() {
	flag.Parse()

	if gIsHelp {
		flag.Usage()
		return
	}

	router()
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(gPort), nil))
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
