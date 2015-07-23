package discovery

import (
	"github.com/coreos/go-etcd/etcd"
	"path"
	"strings"
)

func parse(rawurl string) (scheme string, addrs []string, prefix string) {
	parts := strings.SplitN(rawurl, "://", 2)
	// nodes:port,node2:port => nodes://node1:port,node2:port
	if len(parts) == 1 {
		scheme = "node"
		return
	}
	scheme = parts[0]
	parts = strings.SplitN(parts[1], "/", 2)
	addrs = strings.Split(parts[0], ",")
	if len(parts) == 2 {
		prefix = parts[1]
	}
	return
}

type Store struct {
	root string
	kv   *etcd.Client
}

func makeEndPoints(addrs []string, scheme string) (entries []string) {
	for _, addr := range addrs {
		entries = append(entries, "http"+"://"+addr)
	}
	return
}

func New(rawurl string, root string) (s *Store, err error) {
	_, addrs, prefix := parse(rawurl)
	s = &Store{path.Join(prefix, root), etcd.NewClient(makeEndPoints(addrs, "http"))}
	return
}

func (s *Store) Get(key string, sort, recursive bool) (*etcd.Response, error) {
	return s.kv.Get(path.Join(s.root, key), sort, recursive)
}
