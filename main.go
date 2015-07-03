package main

import (
	"crypto/tls"
	//"encoding/json"
	//"github.com/yihungjen/bigobject-registry"
	"log"
	"net/http"
	"net/url"
	"os"
	"text/template"
)

const (
	// try-bigobject service is of domain trial
	DOMAIN = "trial"

	// try-bigobject provisions user at tier1 machine
	TIER = "tier1"
)

var (
	// Request endpoint multiplexer at PORT 9090
	Server = http.NewServeMux()

	FileServer = http.FileServer(http.Dir("/static"))

	// Command template for our web bosh
	CmdTmpl *template.Template

	// Where our registry is
	Registry *url.URL
)

type MockRequset struct {
	TLS  *tls.ConnectionState
	Host string
}

func HandleRoot(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowd", 405)
		return
	}
	if _, err := r.Cookie("alert-trial-data-volatile"); err != nil {
		http.SetCookie(w, &http.Cookie{
			Name:   "alert-trial-data-volatile",
			Path:   "/",
			MaxAge: 120,
		})
		http.Redirect(w, r, "/alert", 301)
	} else {
		FileServer.ServeHTTP(w, r)
	}
}

func HandleAlert(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "/static/alert.html")
	return
}

func HandleBoshCommand(w http.ResponseWriter, r *http.Request) {
	//if r.Method != "GET" {
	//	http.Error(w, "Method not allowd", 405)
	//	return
	//}
	//headers := w.Header()
	//headers.Add("Content-Type", "application/javascript")

	//payload := url.Values{}
	//payload.Set("domain", DOMAIN)
	//payload.Set("tier", TIER)
	//payload.Set("app", "")

	//resp, err := http.PostForm(Registry.String(), payload)
	//if err != nil {
	//	http.Error(w, "Unable to complete resource acquistion", 503)
	//	return
	//}

	//var resource registry.ResourceEvent
	//if err := json.NewDecoder(resp.Body).Decode(&resource); err != nil {
	//	http.Error(w, "Unable to complete resource acquistion", 503)
	//	return
	//}

	//info := &MockRequset{
	//	TLS:  nil,
	//	Host: resource.Origin,
	//}

	//if err := CmdTmpl.Execute(w, info); err != nil {
	//	log.Println(err)
	//}
	http.NotFound(w, r)
	return
}

func main() {
	s := &http.Server{Addr: ":80", Handler: Server}
	log.Println("begin serving Trial Service...")
	log.Fatalln(s.ListenAndServe())
}

func init() {
	CmdTmpl = template.Must(template.ParseFiles("/static/bosh.command.js"))

	Server.HandleFunc("/", HandleRoot)
	Server.HandleFunc("/alert", HandleAlert)
	Server.HandleFunc("/bosh.command.js", HandleBoshCommand)

	var err error
	Registry, err = url.ParseRequestURI(os.Getenv("REGISTRY_URI"))
	if err != nil {
		log.Fatalln(err)
	}
}
