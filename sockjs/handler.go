package sockjs

import (
	"net/http"
	"regexp"
)

var reInfo = regexp.MustCompile(`^/info$`)
var reIframe = regexp.MustCompile(`^/iframe[\w\d-\. ]*\.html$`)
var reSessionUrl = regexp.MustCompile(
	`^/(?:[\w- ]+)/([\w- ]+)/(xhr|xhr_send|xhr_streaming|eventsource|htmlfile|websocket|jsonp|jsonp_send)$`)
var reRawWebsocket = regexp.MustCompile(`^/websocket$`)

type Handler struct {
	prefix string
	hfunc  func(Session)
	config Config
	pool   *pool
}

func newHandler(pool *pool, prefix string, hfunc func(Session), c Config) (h *Handler) {
	h = new(Handler)
	h.prefix = prefix
	h.hfunc = hfuncCloseWrapper(hfunc)
	h.config = c
	h.pool = pool
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[len(h.prefix):]
	method := r.Method
	println("ServeHTTP:", path, method)

	switch {
	case method == "GET" && path == "" || path == "/":
		greetingHandler(w)
	case method == "GET" && reInfo.MatchString(path):
		infoHandler(h, w, r)
	case method == "OPTIONS" && reInfo.MatchString(path):
		infoOptionsHandler(w, r)
	case method == "GET" && reIframe.MatchString(path):
		iframeHandler(h, w, r)
	case method == "GET" && reRawWebsocket.MatchString(path):
		rawWebsocketHandler(h, w, r)
	case method == "GET" && reSessionUrl.MatchString(path):
		matches := reSessionUrl.FindStringSubmatch(path)
		sessid := matches[1]
		protocol := matches[2]
		switch protocol {
		case "websocket":
			websocketHandler(h, w, r)
		case "eventsource":
			streamingHandler(h, w, r, sessid, eventSourceProtocol{})
		case "htmlfile":
			htmlfileHandler(h, w, r, sessid)
		case "jsonp":
			jsonpHandler(h, w, r, sessid)
		}
	case method == "POST" && reSessionUrl.MatchString(path):
		matches := reSessionUrl.FindStringSubmatch(path)
		sessid := matches[1]
		protocol := matches[2]
		switch protocol {
		case "websocket":
			websocketPostHandler(w, r)
		case "xhr":
			pollingHandler(h, w, r, sessid, xhrPollingProtocol{})
		case "xhr_streaming":
			streamingHandler(h, w, r, sessid, xhrStreamingProtocol{})
		case "xhr_send":
			xhrSendHandler(h, w, r, sessid)
		case "jsonp_send":
			jsonpSendHandler(h, w, r, sessid)
		}
	case method == "OPTIONS" && reSessionUrl.MatchString(path):
		xhrOptionsHandler(w, r)
	default:
		http.NotFound(w, r)
	}
}
