package balancer

import (
	"net/http/httputil"
	"net/url"
	"sync"
)

// Node = singolo server worker
type Node struct {
	URL          *url.URL
	ReverseProxy *httputil.ReverseProxy // inoltra le richieste
	Alive        bool
	Zone         string
	mux          sync.RWMutex // Mutex per evitare problemi di concorrenza
}

// Gestisce la lista dei nodi
type ServerPool struct {
	Nodes []*Node
}