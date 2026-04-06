package emissions

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGracefulDegradationAndCircuitBreaker(t *testing.T) {
    // Variabile per decidere cosa farà il finto server API
    apiShouldFail := false

    // Creazione Server API Mock controllabile
    mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if apiShouldFail {
            // Simula un crollo del server (Errore 500)
            w.WriteHeader(http.StatusInternalServerError)
            return
        }
        // Simula il server che funziona e restituisce 28.0 per la Svezia
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"zone": "SE", "carbonIntensity": 28.0}`))
    }))
    defer mockServer.Close()

    fmt.Println("=== FASE 1: API FUNZIONANTE ===")
    // L'API funziona, ci aspettiamo 28.0 e che la cache si popoli
    val, err := GetCarbonIntensity("SE", "fake-key", mockServer.URL)
    if err != nil {
        t.Fatalf("Errore imprevisto: %v", err)
    }
    fmt.Printf("Valore letto: %v (Atteso: 28.0)\n\n", val)

    fmt.Println("=== FASE 2: API MUORE (SCATTA IL CIRCUIT BREAKER) ===")
    apiShouldFail = true // Spegniamo l'API finta

    // Fa 3 chiamate fallite per far scattare il Circuit Breaker
    for i := 1; i <= 3; i++ {
        fmt.Printf("Tentativo di fallimento %d...\n", i)
        GetCarbonIntensity("SE", "fake-key", mockServer.URL)
    }

    // Alla quarta chiamata, il circuito è aperto. Deve restiruire la STALE CACHE (28.0)
    fmt.Println("-> Test Fallback Livello 1 (Stale Cache)")
    val, _ = GetCarbonIntensity("SE", "fake-key", mockServer.URL)
    fmt.Printf("Valore in emergenza: %v (Atteso: 28.0 dalla cache!)\n\n", val)

    fmt.Println("=== FASE 3: INVECCHIAMENTO CACHE (FALLBACK LIVELLO 2) ===")
    // Invecchia la cache di 4 ore
    cacheMutex.Lock()
    vecchioDato := cache["SE"]
    vecchioDato.LastUpdate = time.Now().Add(-4 * time.Hour)
    cache["SE"] = vecchioDato
    cacheMutex.Unlock()

    // Ora la cache è scaduta. Deve restituire il valore storico per la Svezia (25.5)
    val, _ = GetCarbonIntensity("SE", "fake-key", mockServer.URL)
    fmt.Printf("Valore storico usato: %v (Atteso: 25.5 - Medie storiche!)\n", val)
}