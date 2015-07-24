package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
	"github.com/macrodatalab/try-bigobject/proxy"

	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"text/template"
	"time"
)

const (
	BO_ALERT_KEY = "alert-trial-data-volatile"

	BO_CACHE_TARGET = "bo-trial-target"

	BO_CACHE_EXPIRE = 23 * time.Hour
)

var (
	HostName = os.Getenv("TRIAL_SERVICE_ENDPOINT")

	ServiceImage = os.Getenv("TRIAL_SERVICE_IMAGE")

	DockerHost = os.Getenv("DOCKER_HOST")

	// Request endpoint multiplexer at PORT 9090
	Server = http.NewServeMux()

	FileServer = http.FileServer(http.Dir("/static"))

	// Command template for our web bosh
	CmdTmpl *template.Template
)

func HandleRoot(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowd", 405)
		return
	}

	log.WithFields(log.Fields{"referer": r.Referer(), "url": r.URL}).Debug("GET")

	if _, err := r.Cookie(BO_ALERT_KEY); err != nil {
		http.SetCookie(w, &http.Cookie{
			Name:   BO_ALERT_KEY,
			Value:  "ack",
			Path:   "/",
			MaxAge: 120,
		})
		http.Redirect(w, r, "/alert", 307)
	} else {
		FileServer.ServeHTTP(w, r)
	}

	return
}

func HandleAlert(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "/static/alert.html")
	return
}

type MockRequset struct {
	TLS  *tls.ConnectionState
	Host string
}

func NewInstance() (container *docker.Container, err error) {
	cli, err := docker.NewClient(DockerHost)
	if err != nil {
		return
	}

	container, err = cli.CreateContainer(docker.CreateContainerOptions{
		Config: &docker.Config{Image: ServiceImage},
	})
	if err != nil || container == nil {
		return
	}

	log.WithFields(log.Fields{"container": container.ID}).Info("prepare new instance")

	err = cli.StartContainer(container.ID, &docker.HostConfig{
		PublishAllPorts: true,
	})
	if err != nil {
		return
	}

	return
}

func HandleBoshCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowd", 405)
		return
	}

	// obtain bo identity in cached cookie; if verified go with it, otherwise
	// new instance; if all failed abort
	var container *docker.Container

	iden, err := r.Cookie(BO_CACHE_TARGET)
	if err == nil {
		_, err = proxy.GetNetLoc(iden.Value)
	}

	if err != nil {
		container, err = NewInstance()
		if err != nil {
			log.Error(err)
			http.NotFound(w, r)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:    BO_CACHE_TARGET,
			Value:   container.ID,
			Expires: time.Now().Add(BO_CACHE_EXPIRE),
		})
	} else {
		container = &docker.Container{ID: iden.Value}
	}

	info := &MockRequset{Host: fmt.Sprintf("%s/c/%s", HostName, container.ID)}

	headers := w.Header()
	headers.Add("Content-Type", "application/javascript")

	if err = CmdTmpl.Execute(w, info); err != nil {
		log.Error(err)
		http.NotFound(w, r)
		return
	}

	return
}

func main() {
	CmdTmpl = template.Must(template.ParseFiles("/static/bosh.command.js"))

	Server.Handle("/c/", proxy.NewProxy())

	Server.HandleFunc("/", HandleRoot)
	Server.HandleFunc("/favicon.ico", http.NotFound)
	Server.HandleFunc("/alert", HandleAlert)
	Server.HandleFunc("/bosh.command.js", HandleBoshCommand)

	s := &http.Server{Addr: ":80", Handler: Server}

	log.Println("begin serving Trial Service...")
	log.Fatalln(s.ListenAndServe())
}
