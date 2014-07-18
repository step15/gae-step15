package hello

import (
  "fmt"
  "net/http"
	"strconv"
	"strings"
	"appengine"
	"appengine/urlfetch"
	"net/url"
	"io/ioutil"
	"math/rand"
)

func init() {
  http.HandleFunc("/", root)
  http.HandleFunc("/recv", recv)
  http.HandleFunc("/convert", recv)
	http.HandleFunc("/peers", peers)
	http.HandleFunc("/send", send)
	http.HandleFunc("/show", send)
	http.HandleFunc("/getword", getword)
}

var kPeers = []string {
	"http://1-dot-step-homework-hnoda.appspot.com/stephomeworkhnoda",
	"http://step-test-krispop.appspot.com"}


func root(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, world!")
}

func recv(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	var vs = r.FormValue("content")
	if (len(vs)<1) {
		vs = r.FormValue("message")
	}
	v := strings.Split(vs, "")
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
				fmt.Fprint(w, v[i]);
				i += (k-1) * 2;
			} else if (ik > 0 && ik < k-1) { // middle
				i += ik;
				if (i < len(v)) {
					fmt.Fprint(w, v[i]);
				}
				i += 2 * (k - ik - 1)  // go to bottom and come back
				if (i < len(v)) {
					fmt.Fprint(w, v[i])
				}
				i += ik;  // return to top
			} else { // bottom
				i += ik
				if (i < len(v)) {
					fmt.Fprint(w, v[i])
				}
				i += ik
			}
		}
		if (debug) {
			fmt.Fprint(w, "\n");
		}
	}
}

func peers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, strings.Join(kPeers, "\n"))
}

func send(w http.ResponseWriter, r *http.Request) {
  c := appengine.NewContext(r)
  client := urlfetch.Client(c)
	vs := r.FormValue("message")
	w.Header().Set("Content-Type", "text/plain")
	for i := range kPeers {
		v := url.Values{}
		v.Set("message", vs)
		url := fmt.Sprintf("%s/convert?%s", kPeers[i], v.Encode())
		resp, err := client.Get(url)
		if (err == nil) {
			body, _ := ioutil.ReadAll(resp.Body)
			fmt.Fprintf(w, "%s\n", body)
		} else {
			c.Infof("Error sending to %s => %s", url, err)
		}
	}
}

func getword(w http.ResponseWriter, r *http.Request) {
	verbs := []string {"steam", "bounce", "hop", "jitter"}
	nouns := []string {"banjo", "drum stick", "pine cone", "pretzle"}
	adjectives := []string {"bright", "tasty", "squiggly"}
	animals := []string {"weasel", "unicorn", "dragon", "lemur"}
	
	var word string
	switch(r.FormValue("pos")) {
	case "verb":
		word = PickRandom(verbs)
		break
	case "noun":
		word = PickRandom(append(animals, nouns...))
		break
	case "adjective":
		word = PickRandom(adjectives)
		break
	case "animal":
		word = PickRandom(animals)
		break
	default:
		word = PickRandom(append(append(append(adjectives, animals...), nouns...), verbs...))
	}
	fmt.Fprint(w, word)
}

func PickRandom(choices []string) string {
	return choices[rand.Intn(len(choices))]
}

