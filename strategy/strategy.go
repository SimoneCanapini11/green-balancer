package strategy

import (
	"green-balancer/balancer"
	"sync/atomic"
)

// Interfaccia per tutti gli algoritmi
type Strategy interface {
	NextNode(pool *balancer.ServerPool) *balancer.Node
}

// Algoritmo di base -> RoundRobin
type RoundRobin struct {
	current uint64
}

// Nodo successivo scelto in modo circolare
func (r *RoundRobin) NextNode(pool *balancer.ServerPool) *balancer.Node {
	// Aumenta il contatore di 1 per la concorrenza
	next := atomic.AddUint64(&r.current, uint64(1))
	
	l := uint64(len(pool.Nodes))
	if l == 0 {
		return nil
	}
	
	// Usa il modulo per ricominciare da capo quando finiscono i nodi
	idx := next % l
	return pool.Nodes[idx]
}