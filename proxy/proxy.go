package proxy

import (
	log "github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
	disc "github.com/macrodatalab/try-bigobject/discovery"

	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	FailedAttempts = 3

	ForwardingTarget = 9090
)

var (
	Discovery = os.Getenv("DISCOVERY_HOST")
)

func NewProxy() (s *RegexpHandler) {
	s = &RegexpHandler{}
	s.HandleFunc("/c/[[:xdigit:]]+/cmd", HandleCmdForward)
	s.HandleFunc("/c/[[:xdigit:]]+/(import|info)", HandleEndpoint)
	return
}

func GetIden(key string) (iden string, remain string, err error) {
	parts := strings.SplitN(key, "/", 4)
	if len(parts) != 4 {
		err = fmt.Errorf("bad request URI - %s", key)
		return
	}
	iden = parts[2]
	parts = strings.Split(key, iden)
	remain = parts[1]
	return
}

func GetNetLoc(iden string) (netloc []docker.APIPort, grr error) {
	var retry int
	for {
		dkv, err := disc.New(Discovery, "instances")
		if err != nil {
			log.Fatal("unable to establish connection to discovery -- %v", err)
		}
		resp, err := dkv.Get(iden, false, false)
		if err != nil {
			if retry < FailedAttempts {
				log.Warning(err)
				time.Sleep(1 * time.Second)
				retry += 1
				continue
			} else {
				grr = fmt.Errorf("unable to locate user -- %v", iden)
				break
			}
		}
		err = json.Unmarshal([]byte(resp.Node.Value), &netloc)
		if err != nil {
			grr = err
			return
		}
		break
	}
	return
}

func GetForwardTarget(netloc []docker.APIPort, port int64) (loc *docker.APIPort, err error) {
	for _, nloc := range netloc {
		if nloc.PrivatePort == port {
			return &nloc, nil
		}
	}
	return nil, fmt.Errorf("unable to find target - %d", port)
}

func GetTargetDiscovery(key string) (target *url.URL, err error) {
	iden, remain, err := GetIden(key)
	if err != nil {
		return
	}

	netloc, err := GetNetLoc(iden)
	if err != nil {
		return
	}

	loc, err := GetForwardTarget(netloc, ForwardingTarget)
	if err != nil {
		return
	}

	target = &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", loc.IP, loc.PublicPort),
		Path:   remain,
	}

	log.WithFields(log.Fields{"iden": iden, "url": target}).Info("forward request")

	return
}

func HandleCmdForward(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowd", 405)
		return
	}

	bourl, err := GetTargetDiscovery(r.URL.Path)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), 500)
		return
	}else{
              f, err := os.OpenFile("/log/cmdlog", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
              if err != nil {}
              defer f.Close()
              log.SetOutput(f)
              log.Println(r)

        }
      
       

	resp, err := http.Post(bourl.String(), "application/json", r.Body)
	if err != nil {
		log.Error(err)
		http.Error(w, "internal service error", 500)
		return
	}
	defer resp.Body.Close()
 
	// we reply content in JSON
	w.Header().Add("Content-Type", "application/json")
	io.Copy(NewStreamer(w), resp.Body)

}

func HandleEndpoint(w http.ResponseWriter, r *http.Request) {
	bourl, err := GetTargetDiscovery(r.URL.Path)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), 500)
		return
	}
	r.RequestURI = ""
	r.URL.Scheme = bourl.Scheme
	r.URL.Host = bourl.Host
	r.URL.Path = bourl.Path

	client := &http.Client{}

	// issue request to remote
	resp, err := client.Do(r)
	if err != nil {
		log.Error(err)
		http.Error(w, "internal service error", 500)
		return
	}
	defer resp.Body.Close()

	w.Header().Add("Content-Type", resp.Header.Get("Content-Type"))
	io.Copy(w, resp.Body)
}
