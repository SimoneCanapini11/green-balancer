package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"green-balancer/balancer"
	"green-balancer/config"
	"green-balancer/strategy"
)

func main() {
	fmt.Println("Inizializzazione Green Load Balancer...")

	// Legge il file config.yaml
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Errore critico nel caricamento della configurazione: %v", err)
	}

	// Inizializza il pool e l'algoritmo Round Robin
	pool := &balancer.ServerPool{}
	var loadBalancerStrategy strategy.Strategy = &strategy.RoundRobin{}

	// Trasforma le URL del config.yaml in veri Nodi tramite Reverse Proxy
	for _, nodeCfg := range cfg.Nodes {
		parsedURL, err := url.Parse(nodeCfg.URL)
		if err != nil {
			log.Fatalf("Errore parsing URL %s: %v", nodeCfg.URL, err)
		}

		// Oggetto per inoltro richieste HTTP
		proxy := httputil.NewSingleHostReverseProxy(parsedURL)

		node := &balancer.Node{
			URL:          parsedURL,
			ReverseProxy: proxy,
			Alive:        true,
			Zone:         nodeCfg.Zone,
		}
		pool.Nodes = append(pool.Nodes, node)
	}

	// Creazione server HTTP principale 
	server := http.Server{
		Addr: fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if r.URL.Path == "/favicon.ico" { // Ignora richieste per favicon (icona browser)
				http.NotFound(w, r)
				return
			}
			
			// Usa l'algoritmo per scegliere il nodo a cui inoltrare la richiesta
			targetNode := loadBalancerStrategy.NextNode(pool)
			if targetNode == nil {
				http.Error(w, "Nessun nodo disponibile", http.StatusServiceUnavailable)
				return
			}

			// Stampa dove sta mandando la richiesta
			fmt.Printf("Inoltro richiesta al nodo: %s (Zona: %s)\n", targetNode.URL.String(), targetNode.Zone)
			
			// Inoltra fisicamente la richiesta
			targetNode.ReverseProxy.ServeHTTP(w, r)
		}),
	}

	fmt.Printf("Load Balancer in ascolto sulla porta %d...\n", cfg.Server.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}