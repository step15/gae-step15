package hello

import (
  "fmt"
  "net/http"
	"strconv"
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
	var k, _ = strconv.Atoi(r.FormValue("k"))
	if (k < 1) {
		k = 3
	}
	debug, _ := strconv.ParseBool(r.FormValue("debug"))
	// fmt.Fprintf(w, "k=%d\n", k)
	for ik := 0; ik < k; ik++ {
		if (debug) {
			fmt.Fprintf(w, "%d:", ik);
		}
		for i := 0; i < len(v); {
			if (ik == 0) {
				fmt.Fprint(w, v[i:i+1]);
				i += (k-1) * 2;
			} else if (ik > 0 && ik < k-1) { // middle
				i += ik;
				if (i < len(v)) {
					fmt.Fprint(w, v[i:i+1]);
				}
				i += 2 * (k - ik - 1)  // go to bottom and come back
				if (i < len(v)) {
					fmt.Fprint(w, v[i:i+1])
				}
				i += ik;  // return to top
			} else { // bottom
				i += ik
				if (i < len(v)) {
					fmt.Fprint(w, v[i:i+1])
				}
				i += ik
			}
		}
		if (debug) {
			fmt.Fprint(w, "\n");
		}
	}
}
