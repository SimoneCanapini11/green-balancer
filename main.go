package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"green-balancer/balancer"
	"green-balancer/config"
	"green-balancer/emissions"
	"green-balancer/strategy"
)

func main() {
	fmt.Println("Inizializzazione Green Load Balancer...")

	// Legge il file config.yaml
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Errore critico nel caricamento della configurazione: %v", err)
	}

	// Inizializza il pool e l'algoritmo di bilanciamento
	pool := &balancer.ServerPool{}

	// Strategia basata sul file config.yaml
	var loadBalancerStrategy strategy.Strategy

	if cfg.Balancer.Algorithm == "carbon-aware" {
		fmt.Println("Strategia selezionata: Carbon-Aware")
		loadBalancerStrategy = &strategy.CarbonAware{}
	} else {
		fmt.Println("Strategia selezionata: Round Robin")
		loadBalancerStrategy = &strategy.RoundRobin{}
	}

	// Trasforma le URL del config.yaml in veri Nodi tramite Reverse Proxy
	for _, nodeCfg := range cfg.Nodes {
		parsedURL, err := url.Parse(nodeCfg.URL)
		if err != nil {
			log.Fatalf("Errore parsing URL %s: %v", nodeCfg.URL, err)
		}

		// Oggetto per inoltro richieste HTTP
		proxy := httputil.NewSingleHostReverseProxy(parsedURL)

		node := &balancer.Node{
			URL:            parsedURL,
			ReverseProxy:   proxy,
			Alive:          true,
			Zone:           nodeCfg.Zone,
			CarbonEmission: 9999.0,
			MaxConns:       10,
		}
		pool.Nodes = append(pool.Nodes, node)
	}

	// GOROUTINE: Aggiorna in background le emissioni di carbonio per ogni nodo 
	go func() {
		for {
			for _, node := range pool.Nodes {
				// Chiamata API
				baseURL := cfg.Balancer.APIBaseURL
				intensity, err := emissions.GetCarbonIntensity(node.Zone, cfg.Balancer.APIKey, baseURL)
				if err != nil {
					fmt.Printf("Errore aggiornamento CO2 per zona %s: %v\n", node.Zone, err)
					continue
				}

				// Aggiorna il valore usando il Lock in scrittura
				node.Mux.Lock()
				node.CarbonEmission = intensity
				node.Mux.Unlock()

			}
			// Aspetta 10 secondi prima di controllare di nuovo
			time.Sleep(10 * time.Second) 
		}
	}()

	// GOROUTINE: Controlla in background la salute dei nodi
	go pool.HealthCheck()

	// Creazione server HTTP 
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
				http.Error(w, "Servizio non disponibile. Riprova più tardi.", http.StatusServiceUnavailable)
				return
			}

			// Incrementa il contatore di connessioni attive per il nodo scelto
			targetNode.IncConns()

			// Assicura che il contatore venga decrementato quando la richiesta è completata
			defer targetNode.DecConns() 

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