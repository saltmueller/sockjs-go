// chat room example

package main

import (
	"fmt"
	"github.com/fzzy/sockjs-go/sockjs"
	"net/http"
	"strings"
)

var users *sockjs.Pool = sockjs.NewPool()

func chatHandler(s sockjs.Session) {
	users.Add(s)
	defer users.Remove(s)

	for {
		m := s.Receive()
		if m == nil {
			break
		}
		fullAddr := s.Info().RemoteAddr
		addr := fullAddr[:strings.LastIndex(fullAddr, ":")]
		m = []byte(fmt.Sprintf("%s: %s", addr, m))
		users.Broadcast(m)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/index.html")
}

func main() {
	server := sockjs.NewServer(http.DefaultServeMux)
	conf := sockjs.NewConfig()
	http.Handle("/static", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/", indexHandler)
	server.Handle("/chat", chatHandler, conf)

	err := http.ListenAndServe(":8081", server)
	if err != nil {
		fmt.Println(err)
	}
}