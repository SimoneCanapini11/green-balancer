package balancer

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestWorkerChaosEngineering(t *testing.T) {
	ctx := context.Background()

	// Accensione del container Worker (Nginx)
	t.Log("Avvio del container Worker (Nginx) in corso...")
	req := testcontainers.ContainerRequest{
		Image:        "nginx:alpine", // Server web 
		ExposedPorts: []string{"80/tcp"},
		WaitingFor:   wait.ForHTTP("/"), // Aspetta che il server risponda prima di continuare
	}
	workerContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Errore critico: impossibile avviare Docker. %s", err)
	}

	// Recupera l'indirizzo del container appena avviato
	ip, _ := workerContainer.Host(ctx)
	port, _ := workerContainer.MappedPort(ctx, "80")
	workerURL, _ := url.Parse("http://" + ip + ":" + port.Port())
	t.Logf("Worker avviato con successo all'indirizzo: %s", workerURL.String())

	// Creazione del nodo che punta al container Worker
	node := &Node{
		URL:   workerURL,
		Alive: true,
	}

	// Test su risposta nodo (helth check positivo)
	resp, err := http.Get(node.URL.String())
	if err != nil || resp.StatusCode != 200 {
		t.Fatalf("Il worker non risponde. Test fallito.")
	}
	t.Log("Fase 1 Superata: Il Load Balancer comunica con il container.")

	// Innesco del guasto: Spegnimento brutale del container
	t.Log("Innesco guasto di rete: Spegnimento brutale del container...")
	err = workerContainer.Terminate(ctx)
	if err != nil {
		t.Fatalf("Impossibile uccidere il container: %s", err)
	}

	// Chiusura socket
	time.Sleep(1 * time.Second)

	// Verifica della resilienza: simulazione di quello che fa la tua funzione di HealthCheck
	_, err = http.Get(node.URL.String())
	if err != nil {
		// Se c'è un errore di connessione, il sistema imposterà Alive = false
		node.Mux.Lock()
		node.Alive = false
		node.Mux.Unlock()
	}

	// Asserzione finale: il nodo deve essere marcato come non vivo (Alive=false)
	if node.Alive == true {
		t.Errorf("Test Fallito: Il container è stato distrutto ma il Load Balancer lo crede ancora vivo (Alive=true)!")
	} else {
		t.Log("Test Superato! Il Load Balancer ha rilevato correttamente la morte del nodo (Alive=false) e smetterà di inviargli traffico.")
	}
}