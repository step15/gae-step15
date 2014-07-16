package hello

import (
  "fmt"
  "net/http"
)

func init() {
  http.HandleFunc("/", root)
  http.HandleFunc("/recv", recv)
  http.HandleFunc("/convert", recv)
}

func root(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, world!")
}

func recv(w http.ResponseWriter, r *http.Request) {
	var v = r.FormValue("content")
	if (len(v)<1) {
		v = r.FormValue("message")
	}
	var k = Atoi(r.FormValue("k"))
	if (k < 1) {
		k = 3
	}
	for (int ik = 0; ik < k; ++ik) {
		for (int i = 0; i < len(v);) {
			if (ik == 0) {
				fmt.Print(w, v[i]);
				i += (k-1) * 2;
			} else if (ik > 0 && ik < k-1) { // middle
				i += ik;
				if (i < len(v)) {
					fmt.Fprint(w, v[i]);
				}
				i += 2 * (k - ik)  // go to bottom and come back
				if (i < len(v)) {
					fmt.Fprint(w, v[i])
				}
			} else { // bottom
				i += ik;
				fmt.Fprint(w, [i]);
				i += ik;
			}
		}
	}
}
