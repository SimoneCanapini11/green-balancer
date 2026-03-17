package balancer

import (
	"fmt"
	"net"
	"time"
)

// Prova ad aprire una connessione TCP verso l'indirizzo del nodo
func IsAlive(host string) bool {
	// Prova a connettersi. Se il nodo non risponde entro 2 secondi lo considera morto
	conn, err := net.DialTimeout("tcp", host, 2*time.Second)
	if err != nil {
		return false
	}
	_ = conn.Close() // Se ha risposto, chiudiamo la connessione
	return true
}

// Ciclo che controlla la salute di tutti i nodi
func (s *ServerPool) HealthCheck() {
	for {
		for _, node := range s.Nodes {
			
			host := node.URL.Host 
			isAlive := IsAlive(host)
			node.Mux.Lock()
			
			// Se lo stato è cambiato, stampa avviso per i log
			if node.Alive != isAlive {
				if isAlive {
					fmt.Printf("[HEALTH] Nodo %s ripristinato! (Zona: %s)\n", node.URL.String(), node.Zone)
				} else {
					fmt.Printf("[HEALTH] ATTENZIONE! Nodo %s irraggiungibile! (Zona: %s)\n", node.URL.String(), node.Zone)
				}
			}
			
			node.Alive = isAlive
			node.Mux.Unlock()
		}
		
		// Aspetta 5 secondi prima di fare il prossimo giro di controlli
		time.Sleep(5 * time.Second)
	}
}