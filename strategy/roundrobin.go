package strategy

import (
	"green-balancer/balancer"
	"sync/atomic"
)

// Algoritmo di base -> RoundRobin
type RoundRobin struct {
	current uint64
}

// Nodo successivo scelto in modo circolare
func (r *RoundRobin) NextNode(pool *balancer.ServerPool) *balancer.Node {
	l := uint64(len(pool.Nodes))
	if l == 0 {
		return nil
	}

	// Cerca un nodo vivo
	for range l {
		next := atomic.AddUint64(&r.current, uint64(1))
		idx := next % l
		node := pool.Nodes[idx]

		node.Mux.RLock()
		alive := node.Alive
		node.Mux.RUnlock()

		// Se è vivo, lo restituisce
		if alive {
			return node
		}
	}
	
	// Se il ciclo finisce, tutti i nodi sono morti
	return nil 
}