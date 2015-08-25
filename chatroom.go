package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"golang.org/x/net/websocket"
)

var (
	gPort   int
	gIsHelp bool
)

var gEmotionNums [50]int

func init() {
	for i := 0; i < 50; i++ {
		gEmotionNums[i] = 25 * i
	}

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
	t := template.New("index.html").Delims("<{", "}>")
	t, err := t.ParseFiles("view/index.html")
	if err != nil {
		log.Println("handleIndex:", err)
		http.NotFound(w, r)
		return
	}
	t.Execute(w, map[string]interface{}{
		"emotionNums": gEmotionNums,
	})
}
