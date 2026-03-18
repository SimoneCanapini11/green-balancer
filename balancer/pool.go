package balancer

import (
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic" 
)

// Node = singolo server worker
type Node struct {
	URL            *url.URL
	ReverseProxy   *httputil.ReverseProxy // inoltra le richieste
	Alive          bool
	Zone           string
	CarbonEmission float64 // Emissioni di CO2 per kWh 
	Mux            sync.RWMutex // Mutex per evitare problemi di concorrenza
	ActiveConns    int64 
    MaxConns       int64 
}

// Gestisce la lista dei nodi
type ServerPool struct {
	Nodes []*Node
}

// Incrementa il contatore 
func (n *Node) IncConns() {
  atomic.AddInt64(&n.ActiveConns, 1)
}

// Decrementa il contatore
func (n *Node) DecConns() {
  atomic.AddInt64(&n.ActiveConns, -1)
}

// Legge il valore in tempo reale
func (n *Node) GetConns() int64 {
  return atomic.LoadInt64(&n.ActiveConns)
}