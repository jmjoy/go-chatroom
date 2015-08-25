package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"golang.org/x/net/websocket"
)

var (
	gPort   int  // web server port
	gWsPort int  // websocket server port
	gIsHelp bool // show help info

	gEmotionNums [50]int
)

func init() {
	for i := 0; i < 50; i++ {
		gEmotionNums[i] = i
	}

	flag.IntVar(&gPort, "p", 10000, "web server port")
	flag.IntVar(&gWsPort, "wp", 10001, "websocket server port")
	flag.BoolVar(&gIsHelp, "h", false, "show help")
	flag.BoolVar(&gIsHelp, "help", false, "show help")
}

func main() {
	flag.Parse()

	if gIsHelp {
		flag.Usage()
		return
	}

	wsMux := http.NewServeMux()

	routerWeb()
	routerWebsocket(wsMux)

	go func() {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", gPort), nil))
	}()
	go func() {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", gWsPort), wsMux))
	}()

	select {}
}

func routerWeb() {
	http.HandleFunc("/", handleIndex)
	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("static"))))
}

func routerWebsocket(mux *http.ServeMux) {
	mux.Handle("/ws", websocket.Handler(handleWebsocket))
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
		"wsPort":      gWsPort,
	})
}
