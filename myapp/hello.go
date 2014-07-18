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

var kPeers = []string {
	"http://step-homework-hnoda.appspot.com/",
	"http://step-test-krispop.appspot.com"}

var kMadlibPeers = []string {
	"http://step-test-krispop.appspot.com"}

const kPeersSource = `http://step-homework-hnoda.appspot.com/ T T F F F
http://step-test-krispop.appspot.com/ T T T T T`

var peersServing = initPeers()

func init() {
  http.HandleFunc("/", root)
  http.HandleFunc("/recv", recv)
  http.HandleFunc("/convert", recv)
	http.HandleFunc("/peers", peers)
	http.HandleFunc("/send", send)
	http.HandleFunc("/show", send)
	http.HandleFunc("/getword", getword)
	http.HandleFunc("/madlib", madlib)
}

func initPeers() map[string][]string {
	rMap := make(map[string][]string)
	fields := []string {"url", "convert", "show", "peers", "getword", "madlib"}
	lines := strings.Split(kPeersSource, "\n")
	for li := range(lines) {
		v := strings.Split(lines[li], " ")
//		rMap[lines[li]] = []string {strings.Join(v, ";"),fmt.Sprintf("%d %d", len(v), len(fields))}
		if (len(v) < len(fields)) {
			continue
		}
		url := v[0]
		for fi := 1; fi < len(fields); fi++ {
			val, _ := strconv.ParseBool(v[fi])
//			peerMap[url][fields[fi]] = val
			if (val) {
				rMap[fields[fi]] = append(rMap[fields[fi]], url)
			}
		}
	}
	return rMap
}

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
  c := appengine.NewContext(r)
	ep := r.FormValue("endpoint")
	if (ep == "") {
		ep = "convert"
	}
	fmt.Fprint(w, strings.Join(peersServing[ep], "\n"))
	
}

func send(w http.ResponseWriter, r *http.Request) {
  c := appengine.NewContext(r)
	vs := r.FormValue("message")
	w.Header().Set("Content-Type", "text/plain")

	cs := make(chan string, len(kPeers))
	for i := range kPeers {
		v := url.Values{}
		v.Set("message", vs)
		url := fmt.Sprintf("%s/convert?%s", kPeers[i], v.Encode())
		go FetchUrl(url, c, cs)
	}
	
	for i := range kPeers {
		fmt.Fprintf(w, "%s\n", <-cs)
		i++
	}
}

func FetchUrl(url string, c appengine.Context, cs chan string) {
  client := urlfetch.Client(c)
	c.Infof("Fetching URL: %s", url)
	resp, err := client.Get(url)
	if (err == nil) {
		body, _ := ioutil.ReadAll(resp.Body)
//		c.Infof("Success getting URL: %s", url)
//		cs <- fmt.Sprintf("%s => %s", url, string(body))
		cs <- string(body)
//		c.Infof("Passed channel inject for %s", url)
	} else {
		c.Infof("Error fetching %s => %s", url, err)
		cs <- fmt.Sprintf("[Error: %s]")
	}
}

func getword(w http.ResponseWriter, r *http.Request) {
	verbs := []string {"steam", "bounce", "hop", "jitter"}
	nouns := []string {"wand", "banjo", "drum stick", "pine cone", "pretzle"}
	adjectives := []string {"bright", "tasty", "squiggly", "quixotic"}
	animals := []string {"weasel", "unicorn", "dragon", "lemur"}
	names := []string {"Dexter", "Shiina", "Chrono", "Hermione"}
	adverbs := []string {"furiously", "lazily", "methodically"}
	exclaimations := []string {"Expelliramus", "Lumos", "Expecto patronum", "Wingadium leviosa"}
	
	var word string
	switch(r.FormValue("pos")) {
	case "verb":
		word = PickRandom(verbs)
		break
	case "noun":
		word = PickRandom(nouns)
		break
	case "adjective":
		word = PickRandom(adjectives)
		break
	case "animal":
		word = PickRandom(animals)
		break
	case "name":
		word = PickRandom(names)
		break
	case "exclaimation":
		word = PickRandom(exclaimations)
		break
	case "adverb":
		word = PickRandom(adverbs)
	default:
		word = PickRandom(append(append(append(append(append(adverbs, exclaimations...), adjectives...), animals...), nouns...), verbs...))
	}
	fmt.Fprint(w, word)
}

func PickRandom(choices []string) string {
	return choices[rand.Intn(len(choices))]
}

func GetRandomWord(pos string, c appengine.Context) chan string {
	url := fmt.Sprintf("%s/getword?pos=%s", PickRandom(kMadlibPeers), pos)
	cs := make(chan string, 1)
	FetchUrl(url, c, cs)
	return cs
}	

func madlib(w http.ResponseWriter, r *http.Request) {
  c := appengine.NewContext(r)
	gw := func(pos string) chan string {return GetRandomWord(pos, c)}
	w.Header().Set("Content-Type", "text/plain")
	res := ""
	n := -1
	if (r.FormValue("n")!="") {
		n, _ = strconv.Atoi(r.FormValue("n"))
	}
	if (n < 0 || n > 3) {
		n = rand.Intn(4)
	}
	switch (n) {
	case 0:
		ws := []chan string {gw(""), gw("adjective"), gw("animal"), gw("noun"), gw("adjective")}
		res = fmt.Sprintf(`The mayor of %s-Town was a %s %s. One day the mayor ate a %s and said it was very %s!`, <-ws[0], <-ws[1], <-ws[2], <-ws[3], <-ws[4]);
		break
	case 1:
		ws := []chan string {gw("adjective"), gw("name"), gw("adjective"), gw("animal")}
		res = fmt.Sprintf(`"My what %s teeth you have" said %s the %s %s.`, <-ws[0], <-ws[1], <-ws[2], <-ws[3])
		break
	case 2:
		ws := []chan string {gw("noun"), gw("name"), gw("noun"), gw("animal")}
		res = fmt.Sprintf(`"Now where did I put my %s...?" said %s, while moving the %s before it was eaten by the wild %s.`, <-ws[0], <-ws[1], <-ws[2], <-ws[3])
		break
	case 3:
		ws := []chan string {gw("exclaimation"), gw("name"), gw("adverb"), gw("noun"), gw("adjective"), gw("animal")}
		res = fmt.Sprintf(`"%s!" shouted %s %s, waving her %s at the %s %s.`, <-ws[0], <-ws[1], <-ws[2], <-ws[3], <-ws[4], <-ws[5])
		break
	default:
		panic("Impossible")
	}
	fmt.Fprint(w, res)
}
