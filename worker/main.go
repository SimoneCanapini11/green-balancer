package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	// Porta letta dai comandi del terminale  (9001 di default)
	port := flag.String("port", "9001", "Porta su cui avviare il worker")
	flag.Parse()

	// Crea una singola route che risponde a tutte le richieste
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Simula un tempo di calcolo (es. 100 millisecondi)
		time.Sleep(100 * time.Millisecond)
		
		fmt.Printf("Lavoro eseguito sul worker alla porta %s!\n", *port)
		
		// Risposta inviata al Load Balancer (destinata al browser)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Sono il nodo Worker e sto girando sulla porta %s\n", *port)
	})

	fmt.Printf("Worker in ascolto sulla porta %s...\n", *port)
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		log.Fatal(err)
	}
}