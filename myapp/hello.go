package hello

import (
    "fmt"
    "net/http"
)

func init() {
    http.HandleFunc("/", root)
    http.HandleFunc("/recv", recv)
}

func root(w http.ResponseWriter, r *http.Request) {
    fmt.Fprint(w, "Hello, world!")
}

func recv(w http.ResponseWriter, r *http.Request) {
     fmt.Fprint(w, r.FormValue("content"));
}
