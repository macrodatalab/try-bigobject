package proxy

import (
	"net/http"
)

type Streamer struct {
	w http.ResponseWriter
	f http.Flusher
}

func (s *Streamer) Write(data []byte) (n int, err error) {
	n, err = s.w.Write(data)
	if s.f != nil {
		s.f.Flush()
	}
	return
}

func NewStreamer(w http.ResponseWriter) (s *Streamer) {
	s = &Streamer{w: w}
	if f, ok := w.(http.Flusher); ok {
		s.f = f
	}
	return
}
