import http from 'k6/http';
import { check, sleep } from 'k6';

// Scenario di traffico dinamico
export const options = {
  stages: [
    { duration: '10s', target: 20 }, // Da 0 a 20 utenti in 10 secondi 
    { duration: '20s', target: 20 }, // Mantiene 20 utenti costanti per 20 secondi 
    { duration: '10s', target: 0 },  // Scende gradualmente a 0 utenti in 10 secondi 
  ],
  thresholds: {
    // Il test fallisce se il 95% delle richieste ci mette più di 2 secondi
    http_req_duration: ['p(95)<2000'], 
    // Il test fallisce se c'è anche solo l'1% di errori (Status 500)
    http_req_failed: ['rate<0.01'],   
  },
};

// Comportamento del singolo "Utente Virtuale"
export default function () {
  // URL del Load Balancer 
  const url = 'http://localhost:8080'; 

  const res = http.get(url);

  // Verifica che il Load Balancer abbia risposto con 200 OK
  check(res, {
    'status is 200': (r) => r.status === 200,
  });

  // Pausa per simulare il tempo di lettura (100 millisecondi)
  sleep(0.1); 
}