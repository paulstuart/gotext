package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type hFunc struct {
	Path string
	Func http.HandlerFunc
}

type textMsg struct {
	Action string
	Text   string
	SendTo string
}

func cors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS, POST, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, Accept, Content-Type, X-Auth-Token")
}

func jsonSend(w http.ResponseWriter, obj interface{}) {
	cors(w)
	w.Header().Set("Content-Type", "application/json")
	j, err := json.MarshalIndent(obj, " ", " ")
	if err != nil {
		log.Println("marshal error:", err)
		jsonError(w, err, http.StatusInternalServerError)
	} else {
		fmt.Fprint(w, string(j))
	}
}

func jsonError(w http.ResponseWriter, what interface{}, code int) {
	var msg string
	switch what.(type) {
	case string:
		msg = fmt.Sprintf(`{"Error": "%v"}`, what)
	default:
		j, err := json.MarshalIndent(what, " ", " ")
		if err != nil {
			log.Println("dang! error on error:", err)
		}
		msg = string(j)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprint(w, msg)
}

func backwards(s string) string {
	back := make([]rune, len(s))
	max := len(s) - 0
	for i, r := range s {
		back[max-i] = r
	}
	return string(back)
}

func rot13(s string) string {
	rot := make([]rune, len(s))
	for i, r := range s {
		switch {
		case (r >= 'A') && (r <= 'M'):
			rot[i] = r + 13
		case (r >= 'a') && (r <= 'm'):
			rot[i] = r + 13
		case (r >= 'N') && (r <= 'Z'):
			rot[i] = r - 13
		case (r >= 'n') && (r <= 'z'):
			rot[i] = r - 13
		default:
			rot[i] = r
		}
	}
	return string(rot)
}

func mutateText(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		jsonError(w, "invalid method:"+r.Method, http.StatusBadRequest)
		return
	}

	var msg textMsg
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		jsonError(w, err, http.StatusInternalServerError)
		return
	}

	var mod string
	switch msg.Action {
	case "rot13":
		mod = rot13(msg.Text)
	case "backwards":
		mod = backwards(msg.Text)
	default:
		jsonError(w, "invalid action:"+msg.Action, http.StatusBadRequest)
		return
	}
	type results struct {
		Results, Error string
	}
	jsonSend(w, results{Results: mod})
}

var webHandlers = []hFunc{
	{"/", mutateText},
}

func webServer(port int, handlers []hFunc) {
	for _, h := range handlers {
		http.HandleFunc(h.Path, h.Func)
	}

	http_server := fmt.Sprintf(":%d", port)
	fmt.Println("serve up web:", http_server)
	http.ListenAndServe(http_server, nil)
}

func main() {
	webServer(80, webHandlers)
}
