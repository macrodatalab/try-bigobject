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
	if r.Method != "GET" {
		http.Error(w, "Method not allowd", 405)
		return
	}
	headers := w.Header()
	headers.Add("Content-Type", "application/javascript")

	cli, err := docker.NewClient(DockerHost)
	if err != nil {
		log.Error(err)
		http.NotFound(w, r)
		return
	}

	container, err := cli.CreateContainer(docker.CreateContainerOptions{
		Config: &docker.Config{Image: ServiceImage},
	})
	if err != nil || container == nil {
		log.Error(err)
		http.NotFound(w, r)
		return
	}
	log.WithFields(log.Fields{"container": container.ID}).Info("prepare new instance")

	err = cli.StartContainer(container.ID, &docker.HostConfig{
		PublishAllPorts: true,
	})
	if err != nil {
		log.Error(err)
		http.NotFound(w, r)
		return
	}

	info := &MockRequset{Host: fmt.Sprintf("%s/c/%s", HostName, container.ID)}
	if err = CmdTmpl.Execute(w, info); err != nil {
		log.Error(err)
		http.NotFound(w, r)
		return
	}

	return
}

func main() {
	s := &http.Server{Addr: ":80", Handler: Server}
	log.Println("begin serving Trial Service...")
	log.Fatalln(s.ListenAndServe())
}

func init() {
	CmdTmpl = template.Must(template.ParseFiles("/static/bosh.command.js"))

	Server.Handle("/c/", proxy.NewProxy())

	Server.HandleFunc("/", HandleRoot)
	Server.HandleFunc("/alert", HandleAlert)
	Server.HandleFunc("/bosh.command.js", HandleBoshCommand)
}
