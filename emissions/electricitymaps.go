package emissions

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/sony/gobreaker"
)

// Mappa della struttura del JSON che restituisce l'API
type CarbonResponse struct {
	Zone            string  `json:"zone"`
	CarbonIntensity float64 `json:"carbonIntensity"` // Quantità di CO2 per kWh
}

// Struttura per memorizzare i dati in cache con timestamp
type CachedData struct {
    Intensity  float64
    LastUpdate time.Time
}

var (
    cb         *gobreaker.CircuitBreaker	 // Interruttore per gestire i fallimenti dell'API
    cache      = make(map[string]CachedData) // Cache in memoria per i dati di intensità carbonica
    cacheMutex sync.RWMutex                  // Mutex per sincronizzare l'accesso alla cache
)

// Eseguita in automatico una sola volta all'avvio del programma
func init() {
    settings := gobreaker.Settings{
        Name:        "ElectricityMapsAPI",
        MaxRequests: 1,                // Richieste in stato Half-Open 
        Timeout:     30 * time.Second, // Tempo di attesa quando l'interruttore è Open 
        ReadyToTrip: func(counts gobreaker.Counts) bool {
            // Trip se ci sono 3 o più errori consecutivi 
            return counts.ConsecutiveFailures >= 3
        },
    }
    cb = gobreaker.NewCircuitBreaker(settings)
}

// Funzione helper per il Fail-Closed quando l'API è irraggiungibile o in timeout
func getDefaultPenalty(zone string) float64 {
    switch zone {
    case "SE":
        return 25.5
    case "DE":
        return 350.0
    case "US-TEX-ERCO":
        return 450.0
    default:
        return 500.0 // Peggior scenario per zone sconosciute
    }
}

// Contatta l'API per scoprire quanto inquina un nodo in tempo reale
func GetCarbonIntensity(zone string, apiKey string, baseURL string) (float64, error) {
	
	// FALLBACK: Se manca la chiave API
	if apiKey == "" {
        fmt.Printf("[API MOCK] Lettura simulata per la zona %s...\n", zone)
        return getDefaultPenalty(zone), nil
    }

	// Esecuzione della richiesta attraverso il Circuit Breaker
	result, err := cb.Execute(func() (interface{}, error) {

		// URL per API Electricity Maps
		url := fmt.Sprintf("%s/v3/carbon-intensity/latest?zone=%s", baseURL, zone)

		// Richiesta HTTP in uscita
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return 0, err
		}

		// Aggiunge l'API Key nell'header
		req.Header.Add("auth-token", apiKey)

		// Richiesta eseguita con un Timeout 
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		
		// Gestione errori di rete o timeout
		if err != nil {
			return 0, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return 0, fmt.Errorf("Errore dall'API: status code %d", resp.StatusCode)
		}

		// Risposta JSON -> variabili Go
		var carbonData CarbonResponse
		if err := json.NewDecoder(resp.Body).Decode(&carbonData); err != nil {
			return 0, err
		}

		return carbonData.CarbonIntensity, nil
})	

    // Gestione cache e fallback (Se l'API fallisce o il Circuito è aperto)
    if err != nil {
        fmt.Printf("[CIRCUIT BREAKER] Errore API zona %s. Valuto la Cache...\n", zone)

        // Blocca in lettura per accedere alla mappa in modo sicuro
        cacheMutex.RLock()
        cachedValue, exists := cache[zone]
        cacheMutex.RUnlock()

        // FALLBACK LIVELLO 1: Stale Cache (Dato vecchio ma < 3 ore)
        if exists && time.Since(cachedValue.LastUpdate) < 3*time.Hour {
            tempoTrascorso := time.Since(cachedValue.LastUpdate).Round(time.Second)
            fmt.Printf("[FALLBACK LV1] Uso dati in cache per %s (aggiornati %v fa)\n", zone, tempoTrascorso)
            return cachedValue.Intensity, nil
        }

        // FALLBACK LIVELLO 2: Extreme Degradation (Usa Medie Storiche)
        fmt.Printf("[FALLBACK LV2] Cache scaduta. Uso medie storiche per %s.\n", zone)
        return getDefaultPenalty(zone), nil
    }

	// Se la chiamata ha successo: Aggiorna la cache
    val := result.(float64)

    // Blocca in scrittura per aggiornare la mappa in modo sicuro
    cacheMutex.Lock()
    cache[zone] = CachedData{
        Intensity:  val,
        LastUpdate: time.Now(),
    }
    cacheMutex.Unlock()

    return val, nil
}	