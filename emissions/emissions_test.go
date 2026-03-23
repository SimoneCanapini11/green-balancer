package emissions

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetCarbonIntensity_FailClosed(t *testing.T) {
	// Mock Server che risponde sempre con un Errore 500
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	// Server finto spento alla fine del test
	defer mockServer.Close()

	// Esecuzione della funzione con il server finto e una chiave qualsiasi
	intensity, err := GetCarbonIntensity("SE", "fake-api-key", mockServer.URL)

	// Verifica: la funzione deve catturare l'errore
	if err == nil {
		t.Fatalf("Test Fallito: ci si aspettava un errore dal server, ma non è stato restituito!")
	}

	// Verifica: il valore restituito deve essere 0.0 in caso di errore
	if intensity != 0.0 {
		t.Errorf("Test Fallito: in caso di errore l'intensità restituita deve essere 0.0, ma è %.2f", intensity)
	} else {
		t.Logf("Test Superato! La funzione gestisce correttamente i crash dell'API esterna senza bloccare il programma.")
	}
}