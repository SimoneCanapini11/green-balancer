package strategy

import (
	"green-balancer/balancer"
	"math"
)

// Strategia CarbonAware 
type CarbonAware struct{}

// Sceglie il nodo con le emissioni di carbonio più basse
func (c *CarbonAware) NextNode(pool *balancer.ServerPool) *balancer.Node {
	var bestNode *balancer.Node
	
	// Inizializzazione del minimo
	lowestCarbon := math.MaxFloat64 

	for _, node := range pool.Nodes {
		// Blocchiamo in lettura (evita crash se un'altra funzione sta aggiornando la CO2)
		node.Mux.RLock()
		alive := node.Alive
		carbon := node.CarbonEmission
		max := node.MaxConns
		node.Mux.RUnlock()

		// Connessioni attive in questo momento
   		active := node.GetConns()

		// Sceglie un nodo se: è vivo && inquina meno && non è saturo
        if alive && carbon < lowestCarbon && active < max {
        	bestNode = node
      		lowestCarbon = carbon
    	}
	}

	return bestNode
}