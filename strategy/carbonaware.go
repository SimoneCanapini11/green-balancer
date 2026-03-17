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
	
	// Inizializzazione del minimo con un numero molto alto
	lowestCarbon := math.MaxFloat64 

	for _, node := range pool.Nodes {
		// Blocchiamo in lettura (evita crash se un'altra funzione sta aggiornando la CO2)
		node.Mux.RLock()
		alive := node.Alive
		carbon := node.CarbonEmission
		node.Mux.RUnlock()

		// Se il nodo è vivo e inquina meno del minimo attuale...
		if alive && carbon < lowestCarbon {
			lowestCarbon = carbon
			bestNode = node // ...diventa lui il nuovo candidato ideale!
		}
	}

	return bestNode
}