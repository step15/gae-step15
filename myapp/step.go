package step

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"appengine"
	"appengine/datastore"
	"appengine/memcache"
	"appengine/urlfetch"
)

const kPeerStoreKind = "peerSouce"
const kPeerStoreId = "current"
const kSelfName = "step15-krispop"
const kSelfUrl = "http://" + kSelfName + ".appspot.com"
const kPeerSourceStatic = `http://step15-krispop.appspot.com/	T	T	T	T	T
http://regal-sun-100211.appspot.com	T	T	T	T	T`

func init() {
	http.HandleFunc("/", root)
	http.HandleFunc("/recv", recv)
	http.HandleFunc("/convert", recv)
	http.HandleFunc("/peers", peers)
	http.HandleFunc("/send", send)
	http.HandleFunc("/show", send)
	http.HandleFunc("/getword", getword)
	http.HandleFunc("/madlib", madlib)
	http.HandleFunc("/update-peers", updatePeers)
}

var peerSplitRe = regexp.MustCompile(`\t`)
var trailingSlashRe = regexp.MustCompile("/$")
var appspotPrefixRe = regexp.MustCompile(`\.appspot.com.*`)
var appspotMatchRe = regexp.MustCompile(`http://([^."']+)\.appspot\.com[^"']*`)

type FetchRes struct {
	Url, Res string
}

type StringStruct struct {
	s string
}

type SimpleMessage struct {
	Result string `json:"result"`
}

type PeersMessage struct {
	Peers []string `json:"peers"`
}

type ShowMessage struct {
	ShowResults []FetchRes `json:"showResults"`
}

func initPeers(c appengine.Context) map[string][]string {
	rMap := make(map[string][]string)
	fields := []string{"url", "convert", "show", "getword", "madlib", "peers"}
	lines := strings.Split(GetPeers(c), "\n")
	for li := range lines {
		v := peerSplitRe.Split(lines[li], len(fields))
		for len(v) < len(fields) {
			v = append(v, "F")
		}
		//		log.Printf("Got %s", strings.Join(v, ";"))
		url := trailingSlashRe.ReplaceAllString(v[0], "")
		if !appspotMatchRe.MatchString(url) {
			continue
		}
		for fi := 1; fi < len(fields); fi++ {
			val, _ := strconv.ParseBool(v[fi])
			if val {
				rMap[fields[fi]] = append(rMap[fields[fi]], url)
			}
		}
	}
	log.Printf("Peers map: %s", rMap)
	return rMap
}

func allPeers(c appengine.Context) []string {
	m := make(map[string]bool)
	r := []string{}
	for _, v := range initPeers(c) {
		for _, peer := range v {
			m[peer] = true
		}
	}
	for k, _ := range m {
		r = append(r, k)
	}
	sort.Strings(r)
	return r
}

func contains(needle string, haystack []string) bool {
	for _, x := range haystack {
		if needle == x {
			return true
		}
	}
	return false
}

func selfBase(r *http.Request) string {
	appengine.NewContext(r).Infof("request = %v", r)
	host := kSelfUrl
	if server, port := r.Header["X-Appengine-Server-Name"], r.Header["X-Appengine-Server-Port"]; len(server) > 0 {
		host = server[0]
		if len(port) > 0 {
			host = host + ":" + port[0]
		}
	}
	return fmt.Sprintf("http://%s", host)
}

func root(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	baseline := ""
	c := appengine.NewContext(r)
	peers := allPeers(c)
	base := r.FormValue("base")
	if base != "" && contains(base, peers) {
		baseline = "<base href=" + base + ">"
	}

	// make base selector magic
	baseSelector := `<form method=get action='` + selfBase(r) + `/'>
    <select name=base onchange="this.form.submit()">`
	for _, peer := range peers {
		names := appspotMatchRe.FindStringSubmatch(peer)
		name := peer
		if len(names) > 1 {
			name = names[1]
		}
		c.Infof("peer:%v => names:%v => name:%v", peer, names, name)
		selected := ""
		if len(base) < 1 || base == "http://localhost:8080" {
			base = kSelfUrl
		}
		if peer == base {
			selected = "selected"
		}
		baseSelector += `<option ` + selected + ` value="` + peer + `">` + name + `</option>`
	}
	baseSelector += `</select>で実行</form>`

	fmt.Fprint(w, `
<head>
<link rel="stylesheet" href="pure/pure-min.css">
<title>STEP HW7 例文のサーバー</title>
`+baseline+`
<style>
th {
  background-color: #e0e0e0;
  color: #000;
}
td {
  vertical-align: middle;
}
td.left {
  text-align: start;
  vertical-align: middle;
}
td.right {
  text-align: right;
  vertical-align: middle;
}
form {
  vertical-align: middle;
  margin: 0;
}
</style>
</head>
<body style="margin: 10px">
<table border=1 class=pure-table-striped>
<thead><tr>
<th width=50%>以下のリンクを試してください</th>
<th width=50% class=right>フォームに何かを入力したら、飛ぶ先のURLを見てみてください</th>
</tr></thead><tbody>
<tr><td><a href="/convert?message=DoAndroidsDreamOfElectricSheep?">/convert?message=DoAndroidsDreamOfElectricSheep?</a>
  <td class=right><form method=get action=convert><input placeholder="このinputタグの名前が「message」です" type=text name=message size=42><input type=submit value="Convert!"></form></tr>
<tr><td><a href="/show?message=ピタゴラ暗号棒ではない">/show?message=ピタゴラ暗号棒ではない</a>
  <td class=right><form method=get action=show><input placeholder="このinputタグの名前が「message」です" type=text name=message size=42><input type=submit value="Show!"></form></tr>
<tr><td><a href="/peers">/peers</a> (表示されるサーバーは/convertに対応しています)<br>
  <a href="/peers?endpoint=getword">/peers?endpoint=getword</a> (表示されるサーバーは/getwordを対応しているので、/madlibを作るのに使えます)
  <td class=right><form method=get action=peers>
    <select name=endpoint>
    <option value="convert">convert</option>
    <option value="show">show</option>
    <option value="getword">getword</option>
    <option value="madlib">madlib</option>
    <option value="peers">peers</option>
    </select>
    <input type=submit value="Peers!">
    </form></tr>
<tr><td><a href="/getword?pos=animal">/getword?pos=animal</a> このサーバーは以下のparts of speech (pos)に対応しています: verb, noun, adjective, animal, name, adverb, exclaimation. 好きなposを実装してください. 対応していないposのリクエストがをもらったら適当なランダムの単語を返しても構いません！
  <td class=right><form method=get action=getword>
    <select name=pos>
    <option value=""></option>
    <option value="verb">verb</option>
    <option value="noun">noun</option>
    <option value="adjective">adjective</option>
    <option value="adverb">adverb</option>
    <option value="animal">animal</option>
    <option value="name">name</option>
    <option value="exclaimation">exclaimation</option>
    </select>
    <input type=submit value="Getword!"></form>
<tr><td><a href="/madlib">/madlib</a> (ランダムなmadlibを作成します)
  <td class=right><form method=get action=madlib><input type=submit value="Madlib!"></form></tr>
</tbody>
</table>
`+baseSelector+`
<div style="height:100em"></div>
ちなみに、私の/convertのURLは何をやっているんだろうと不思議に思ってたら、
これで試してみてください：
<form method=get action=convert>
<table class="pure-table pure-table-bordered">
<tr><td class=right>message<td><input type=text name=message size=100 value="Life is like riding a bicycle. To keep your balance you must keep moving.
  --Albert Einstein">
<tr><td class=right>k<td><input type=text name=k size=2 value=5>
<tr><td class=right>magic<td><input type=checkbox name=debug checked=true>
</table>
<input type=submit value=convert>
</form>

</body>
`)
}

func AddHeaders(w *http.ResponseWriter) {
	(*w).Header().Set("Content-Type", "text/plain; charset=utf-8")
}

func ReqWantsJson(r *http.Request) bool {
	if r.FormValue("fmt") == "json" {
		return true
	}
	v, _ := strconv.ParseBool(r.FormValue("json"))
	return v
}

func RailCipher(vs string, k int, debug bool) string {
	var w bytes.Buffer

	v := strings.Split(vs, "")
	for ik := 0; ik < k; ik++ {
		if ik >= len(v) {
			break
		}
		if debug {
			if ik > 0 {
				fmt.Fprint(&w, "\n")
			}
			fmt.Fprintf(&w, "%2d:", ik)
		}
		for i := 0; i < len(v); {
			advance := func(x, printed int) {
				if debug {
					fmt.Fprint(&w, strings.Repeat("-", x-printed))
				}
				i += x
			}

			if ik == 0 {
				fmt.Fprint(&w, v[i])
				advance((k-1)*2, 1)
			} else if ik > 0 && ik < k-1 { // middle
				advance(ik, 0)
				if i < len(v) {
					fmt.Fprint(&w, v[i])
				}
				advance(2*(k-ik-1), 1) // go to bottom and come back
				if i < len(v) {
					fmt.Fprint(&w, v[i])
				}
				advance(ik, 1) // return to top
			} else { // bottom
				advance(ik, 0)
				if i < len(v) {
					fmt.Fprint(&w, v[i])
				}
				advance(ik, 1)
			}
		}
	}
	return w.String()
}

func recv(w http.ResponseWriter, r *http.Request) {
	//	c := appengine.NewContext(r)
	AddHeaders(&w)
	debug, _ := strconv.ParseBool(r.FormValue("debug"))
	switch r.FormValue("debug") {
	case "on":
		debug = true
	}
	var vs = r.FormValue("content")
	if len(vs) < 1 {
		vs = r.FormValue("message")
	}
	var k, _ = strconv.Atoi(r.FormValue("k"))
	if k < 1 {
		k = 3
	}
	es := RailCipher(vs, k, debug)
	SimpleResponse(w, r, es)
}

func SimpleResponse(w io.Writer, r *http.Request, s string) {
	var rm SimpleMessage
	rm.Result = s
	if ReqWantsJson(r) {
		js, _ := json.MarshalIndent(rm, "", "  ")
		fmt.Fprint(w, string(js))
	} else {
		fmt.Fprint(w, rm.Result)
	}
}

func peers(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	AddHeaders(&w)
	ep := r.FormValue("endpoint")
	if ep == "" {
		ep = "convert"
	}
	peers := initPeers(c)
	fmt.Fprint(w, strings.Join(peers[ep], "\n"))
}

func send(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	vs := r.FormValue("message")
	AddHeaders(&w)
	kPeers := initPeers(c)["convert"]

	cf := make(chan FetchRes, len(kPeers))
	for i := range kPeers {
		v := url.Values{}
		v.Set("message", vs)
		url := fmt.Sprintf("%s/convert?%s", kPeers[i], v.Encode())
		go FetchUrl(url, c, cf)
	}

	var rm ShowMessage
	for _ = range kPeers {
		rm.ShowResults = append(rm.ShowResults, <-cf)
	}

	if ReqWantsJson(r) {
		rs, _ := json.MarshalIndent(rm, "", "  ")
		fmt.Fprint(w, string(rs))
	} else {
		for _, res := range rm.ShowResults {
			showUrl := appspotPrefixRe.ReplaceAllString(res.Url, "...")
			showRes := strings.Replace(strings.TrimSpace(res.Res), "\n", " ↓ ", -1)
			fmt.Fprintf(w, "%s => %s\n", showUrl, showRes)
		}
	}
}

func FetchUrl(url string, c appengine.Context, cf chan FetchRes) {
	client := urlfetch.Client(c)
	c.Infof("Fetching URL: %s", url)
	resp, err := client.Get(url)
	var r FetchRes
	r.Url = url
	if err == nil {
		body, _ := ioutil.ReadAll(resp.Body)
		r.Res = string(body)
		if resp.StatusCode != 200 {
			r.Res = fmt.Sprintf("[Failure (%s): %s]", resp.Status, r.Res)
		}
		//		c.Infof("Success getting URL: %s => %s", url, r.Res)
	} else {
		c.Warningf("Error fetching %s => %s", url, err)
		r.Res = fmt.Sprintf("[Error: %s]", err)
	}
	cf <- r
}

func getword(w http.ResponseWriter, r *http.Request) {
	verbs := []string{"steam", "bounce", "hop", "jitter"}
	nouns := []string{"wand", "banjo", "drum stick", "pine cone", "pretzle"}
	adjectives := []string{"bright", "tasty", "squiggly", "quixotic"}
	animals := []string{"weasel", "unicorn", "dragon", "lemur"}
	names := []string{"Dexter", "Shiina", "Chrono", "Hermione"}
	adverbs := []string{"furiously", "lazily", "methodically"}
	exclaimations := []string{"Expelliramus", "Lumos", "Expecto patronum", "Wingadium leviosa"}

	var word string
	switch r.FormValue("pos") {
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
	SimpleResponse(w, r, word)
}

func PickRandom(choices []string) string {
	return choices[rand.Intn(len(choices))]
}

func GetRandomWord(pos string, c appengine.Context) chan string {
	kPeers := initPeers(c)["getword"]
	url := fmt.Sprintf("%s/getword?pos=%s", PickRandom(kPeers), pos)
	cs := make(chan string, 1)
	go func() {
		cf := make(chan FetchRes, 1)
		FetchUrl(url, c, cf)
		p := <-cf
		cs <- p.Res
	}()
	return cs
}

func madlib(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	gw := func(pos string) chan string { return GetRandomWord(pos, c) }
	AddHeaders(&w)
	res := ""
	n := -1
	if r.FormValue("n") != "" {
		n, _ = strconv.Atoi(r.FormValue("n"))
	}
	if n < 0 || n > 3 {
		n = rand.Intn(4)
	}
	switch n {
	case 0:
		ws := []chan string{gw(""), gw("adjective"), gw("animal"), gw("noun"), gw("adjective")}
		res = fmt.Sprintf(`The mayor of %s-Town was a %s %s. One day the mayor ate a %s and said it was very %s!`, <-ws[0], <-ws[1], <-ws[2], <-ws[3], <-ws[4])
		break
	case 1:
		ws := []chan string{gw("adjective"), gw("name"), gw("adjective"), gw("animal")}
		res = fmt.Sprintf(`"My what %s teeth you have" said %s the %s %s.`, <-ws[0], <-ws[1], <-ws[2], <-ws[3])
		break
	case 2:
		ws := []chan string{gw("noun"), gw("name"), gw("noun"), gw("animal")}
		res = fmt.Sprintf(`"Now where did I put my %s...?" said %s, while moving the %s before it was eaten by the wild %s.`, <-ws[0], <-ws[1], <-ws[2], <-ws[3])
		break
	case 3:
		ws := []chan string{gw("exclaimation"), gw("name"), gw("adverb"), gw("noun"), gw("adjective"), gw("animal")}
		res = fmt.Sprintf(`"%s!" shouted %s %s, waving her %s at the %s %s.`, <-ws[0], <-ws[1], <-ws[2], <-ws[3], <-ws[4], <-ws[5])
		break
	default:
		panic("Impossible")
	}
	SimpleResponse(w, r, res)
}

func StorePeers(c appengine.Context, s string) {
	ss := new(StringStruct)
	ss.s = s
	key := datastore.NewKey(c, kPeerStoreKind, kPeerStoreId, 0, nil)
	_, err := datastore.Put(c, key, ss)
	if err != nil {
		panic(err)
	}
	StorePeersCached(c, s)
}

func GetPeersCached(c appengine.Context) string {
	item, err := memcache.Get(c, kPeerStoreKind)
	if err == memcache.ErrCacheMiss {
		return ""
	} else if err != nil {
		c.Errorf("error getting item: %v", err)
	}
	return string(item.Value)
}

func StorePeersCached(c appengine.Context, s string) {
	item := &memcache.Item{
		Key:   kPeerStoreKind,
		Value: []byte(s),
	}
	err := memcache.Set(c, item)
	if err != nil {
		c.Errorf("error setting item: %v", err)
	}
}

func GetPeers(c appengine.Context) string {
	if c == nil {
		return kPeerSourceStatic
	}
	var s string
	s = GetPeersCached(c)
	if s != "" {
		return s
	}
	ss := new(StringStruct)
	key := datastore.NewKey(c, kPeerStoreKind, kPeerStoreId, 0, nil)
	err := datastore.Get(c, key, ss)
	if err != nil {
		c.Errorf("Error reading from datastore %s", err)
	}
	s = ss.s
	if s == "" {
		s = kPeerSourceStatic
		c.Warningf("Datastore read failed. Using static peers")
	}
	StorePeersCached(c, s)
	return s
}

func updatePeers(w http.ResponseWriter, r *http.Request) {
	contentB, _ := ioutil.ReadAll(r.Body)
	content := string(contentB)
	c := appengine.NewContext(r)

	AddHeaders(&w)

	if len(content) < 5 {
		c.Errorf("Empty update peer request. Rejected.")
		fmt.Fprint(w, "Emtpy request rejected")
		return
	}

	if len(content) > 9000 {
		c.Errorf("Update peer request! It's over 9000! Rejectzored!")
		fmt.Fprint(w, "Too long. Rejectzored.")
		return
	}

	old := GetPeersCached(c)
	if old == content {
		fmt.Fprint(w, "Same as existing. Ignored.")
		return
	}

	c.Infof("Got new update: %s", content)
	StorePeers(c, content)

	fmt.Fprint(w, "Updated")
}
