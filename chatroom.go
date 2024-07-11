package main

import (
	"flag"
	"fmt"
	"github.com/jmjoy/go-chatroom/static"
	"github.com/jmjoy/go-chatroom/view"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"golang.org/x/net/websocket"
)

const (
	VERSION = "v0.1"

	DEBUG = true
)

var (
	gPort         int // web server port
	gWsPort       int // websocket server port
	gClientWsPort int // websocket server port using in browser

	gIsHelp bool // show help info

	gEmotionNums [50]int

	//gRegEscape = regexp.MustCompile(`<script[\s\S]*?>[\s\S]*?</script>`)

	gUploadDir string // upload file dir
)

func init() {
	for i := 0; i < 50; i++ {
		gEmotionNums[i] = i
	}

	flag.IntVar(&gPort, "p", 10000, "web server port")
	flag.IntVar(&gWsPort, "wp", 10001, "websocket server port")
	flag.IntVar(&gClientWsPort, "cwp", 0, "websocket server port using in browser, same as wp if zero")
	flag.BoolVar(&gIsHelp, "h", false, "show help")
	flag.BoolVar(&gIsHelp, "help", false, "show help")
	flag.StringVar(&gUploadDir, "ud", "./go-chatroom-upload", "upload file path")
}

func main() {
	flag.Parse()

	if gIsHelp {
		flag.Usage()
		return
	}

	// Create upload dir if not exists
	if _, err := os.Stat(gUploadDir); os.IsNotExist(err) {
		if err := os.Mkdir(gUploadDir, 0755); err != nil {
			log.Fatal(err)
		}
	} else if err != nil {
		log.Fatal(err)
	}

	if gClientWsPort == 0 {
		gClientWsPort = gWsPort
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
	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.FS(static.FS))))
	http.Handle("/upload/", http.StripPrefix("/upload", http.FileServer(http.Dir(gUploadDir))))
}

func routerWebsocket(mux *http.ServeMux) {
	mux.Handle("/ws", websocket.Handler(handleWebsocket))
}

var gFuncMap = template.FuncMap{
	"op": operate,
}

func operate(op string, a, b int) string {
	var result int

	switch op {
	case "+":
		result = a + b
	case "-":
		result = a - b
	case "*":
		result = a * b
	case "/":
		result = a / b
	}

	return strconv.Itoa(result)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "/" {
		t := template.New("index.html").Delims("<{", "}>").Funcs(gFuncMap)
		t, err := t.Parse(view.IndexHtml)
		if err != nil {
			log.Println("handleIndex:", err)
			http.NotFound(w, r)
			return
		}
		t.Execute(w, map[string]interface{}{
			"emotionNums": gEmotionNums,
			"wsPort":      gClientWsPort,
		})
	} else {
		http.NotFound(w, r)
	}
}

func logError(v interface{}) {
	if DEBUG {
		panic(v)
		return
	}
	log.Println(v)
}

//@Deprecated
//func escapeBody(body []byte) []byte {
//    indexss := gRegEscape.FindAllIndex(body, -1)
//    if len(indexss) == 0 {
//        return body
//    }
//    var buffer bytes.Buffer
//    var i int
//    for _, indexs := range indexss {
//        fmt.Println(i, indexs[0], indexs[1])
//        buffer.Write(body[i:indexs[0]])
//        buffer.WriteString(html.EscapeString(string(body[indexs[0]:indexs[1]])))
//        i = indexs[1]
//    }
//    return buffer.Bytes()
//}
