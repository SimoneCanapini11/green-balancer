package strategy

import (
	"net/url"
	"testing"

	"green-balancer/balancer" 
)

func TestCarbonAwareSpillover(t *testing.T) {
	// URL per i nodi
	url1, _ := url.Parse("http://worker-svezia:9001")
	url2, _ := url.Parse("http://worker-texas:9002")
	url3, _ := url.Parse("http://worker-germania:9003")

	// Pool di server 
	pool := &balancer.ServerPool{
		Nodes: []*balancer.Node{
			{
				URL:            url1,
				Zone:           "SE",
				Alive:          true,
				CarbonEmission: 20.0, 
				MaxConns:       10,
				ActiveConns:    10,   
			},
			{
				URL:            url2,
				Zone:           "US-TEX-ERCO",
				Alive:          true,
				CarbonEmission: 200.0, 
				MaxConns:       10,
				ActiveConns:    5,     
			},
			{
				URL:            url3,
				Zone:           "DE",
				Alive:          true,
				CarbonEmission: 350.0, 
				MaxConns:       10,
				ActiveConns:    0,     
			},
		},
	}

	// Inizializzazione strategia
	caStrategy := &CarbonAware{}

	// Esecuzione funzione da testare
	selectedNode := caStrategy.NextNode(pool)

	// Verifica: L'algoritmo deve scegliere il Texas (url2)
	if selectedNode == nil {
		t.Fatalf("Errore: L'algoritmo non ha restituito nessun nodo")
	}

	if selectedNode.URL.String() != url2.String() {
		t.Errorf("Test Fallito! Ci si aspettava il nodo Texas (%s), ma l'algoritmo ha scelto: %s", url2.String(), selectedNode.URL.String())
	} else {
		t.Logf("Test Superato! L'algoritmo funziona: la Svezia era satura, quindi ha deviato correttamente sul Texas.")
	}
}

func TestCarbonAwareBase(t *testing.T) {
	// URL per i nodi
	url1, _ := url.Parse("http://worker-svezia:9001")
	url2, _ := url.Parse("http://worker-texas:9002")
	url3, _ := url.Parse("http://worker-germania:9003")

	// Pool di server 
	pool := &balancer.ServerPool{
		Nodes: []*balancer.Node{
			{
				URL:            url1,
				Zone:           "SE",
				Alive:          true,
				CarbonEmission: 20.0, 
				MaxConns:       10,
				ActiveConns:    0,    
			},
			{
				URL:            url2,
				Zone:           "US-TEX-ERCO",
				Alive:          true,
				CarbonEmission: 200.0,
				MaxConns:       10,
				ActiveConns:    0,    
			},
			{
				URL:            url3,
				Zone:           "DE",
				Alive:          true,
				CarbonEmission: 350.0,
				MaxConns:       10,
				ActiveConns:    0,    
			},
		},
	}

	caStrategy := &CarbonAware{}
	selectedNode := caStrategy.NextNode(pool)

	if selectedNode == nil {
		t.Fatalf("Errore: L'algoritmo non ha restituito nessun nodo")
	}

	// Verifica: Essendo tutti vuoti, deve scegliere la Svezia (url1)
	if selectedNode.URL.String() != url1.String() {
		t.Errorf("Test Fallito! Ci si aspettava la Svezia (%s), ma ha scelto: %s", url1.String(), selectedNode.URL.String())
	} else {
		t.Logf("Test Superato! L'algoritmo base funziona: ha scelto la Svezia perché è il nodo più verde e ha posti liberi.")
	}
}

func TestCarbonAwareFailoverAndRecovery(t *testing.T) {
	// URL per i nodi
	url1, _ := url.Parse("http://worker-svezia:9001")
	url2, _ := url.Parse("http://worker-germania:9003")

	node1 := &balancer.Node{
		URL:            url1,
		Zone:           "SE",
		Alive:          true, 
		CarbonEmission: 20.0, 
		MaxConns:       10,
		ActiveConns:    0,
	}
	
	node2 := &balancer.Node{
		URL:            url2,
		Zone:           "DE",
		Alive:          true,
		CarbonEmission: 350.0, 
		MaxConns:       10,
		ActiveConns:    0,
	}

	pool := &balancer.ServerPool{
		Nodes: []*balancer.Node{node1, node2},
	}

	caStrategy := &CarbonAware{}

	// Fase1: Entrambi i nodi sono online, deve scegliere la Svezia
	t.Log("Fase 1: Entrambi i nodi sono online.")
	res1 := caStrategy.NextNode(pool)
	if res1.URL.String() != url1.String() {
		t.Fatalf("Errore Fase 1: Ci si aspettava Svezia, ma ha scelto %s", res1.URL.String())
	}

	// Fase2 (Failover): La Svezia va offline, deve scegliere la Germania 
	t.Log("Fase 2: Il nodo Svezia va offline (Alive = false).")
	node1.Mux.Lock()
	node1.Alive = false // Simuliamo che l'HealthCheck lo abbia dichiarato morto
	node1.Mux.Unlock()

	res2 := caStrategy.NextNode(pool)
	if res2.URL.String() != url2.String() {
		t.Fatalf("Errore Fase 2: Ci si aspettava Germania (Failover), ma ha scelto %s", res2.URL.String())
	}

	// Fase3 (Recovery): La Svezia torna online, deve scegliere di nuovo la Svezia
	t.Log("Fase 3: Il nodo Svezia viene riavviato e torna online (Alive = true).")
	node1.Mux.Lock()
	node1.Alive = true // Simuliamo che l'HealthCheck lo veda di nuovo vivo
	node1.Mux.Unlock()

	res3 := caStrategy.NextNode(pool)
	if res3.URL.String() != url1.String() {
		t.Fatalf("Errore Fase 3: Ci si aspettava il ritorno alla Svezia (Recovery), ma ha scelto %s", res3.URL.String())
	}

	t.Log("Test Superato! Il sistema gestisce perfettamente il Failover e il Recovery automatico (Self-Healing).")
}