package emissions

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Mappa della struttura del JSON che restituisce l'API
type CarbonResponse struct {
	Zone            string  `json:"zone"`
	CarbonIntensity float64 `json:"carbonIntensity"` // Quantità di CO2 per kWh
}

// Contatta l'API per scoprire quanto inquina un nodo in tempo reale
func GetCarbonIntensity(zone string, apiKey string) (float64, error) {
	
	// FALLBACK: Se non hai ancora inserito la chiave, usiamo dati simulati realistici -----------
	if apiKey == "CHIAVE_DA_INSEIRE" || apiKey == "" {
		fmt.Printf("[API MOCK] Lettura simulata per la zona %s...\n", zone)
		mockValues := map[string]float64{
			"SE":    20.5,  // Svezia: molto verde 
			"DE":    350.0, // Germania: inquina di più 
			"US-TEX-ERCO": 450.0, // Texas: inquina molto
		}
		if val, exists := mockValues[zone]; exists {
			return val, nil
		}
		return 500.0, nil // Valore pessimo di default se la zona è sconosciuta
	}

	// Indirizzo Free Tier di Electricity Maps
	url := fmt.Sprintf("https://api.electricitymap.org/v3/carbon-intensity/latest?zone=%s", zone)

	// Richiesta HTTP in uscita
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	// Aggiunge l'API Key nell'header
	req.Header.Add("auth-token", apiKey)

	// Richiesta eseguita con un Timeout 
	// Se l'API ci mette più di 5 secondi a rispondere, annulliamo tutto per non bloccarci.
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("errore dall'API: status code %d", resp.StatusCode)
	}

	// Risposta JSON -> variabili Go
	var carbonData CarbonResponse
	if err := json.NewDecoder(resp.Body).Decode(&carbonData); err != nil {
		return 0, err
	}

	return carbonData.CarbonIntensity, nil
}