# Carbon-Aware Load Balancer

Questo repository contiene l'implementazione di un Reverse Proxy sviluppato in **Go**, progettato per integrare i principi del *Green Cloud Computing* nel routing di rete. 

Il sistema è dotato di logiche di instradamento dinamico **Carbon-Aware**, bilanciando il traffico verso le regioni Cloud con la minore intensità carbonica in tempo reale. Per garantire High Availability e prevenire fallimenti a cascata, l'architettura implementa pattern di resilienza avanzati, tra cui **Circuit Breaker** e **Graceful Degradation**. L'infrastruttura è stata collaudata sia in ambiente locale tramite container, sia in produzione su istanze Amazon Web Services (AWS EC2).

---

## 1. Prerequisiti

Per eseguire l'architettura in locale e lanciare le suite di test, è necessario aver installato sulla propria macchina:
- **Docker** e **Docker Compose** (per l'orchestrazione dei container)
- **Go** (versione 1.20 o superiore, per eseguire gli Unit Test isolati)
- **k6** (framework open-source di Grafana Labs, per gli Stress Test)

---

## 2. Configurazione Iniziale

Il bilanciatore interroga in tempo reale le API di telemetria (es. Electricity Maps) per ottenere l'intensità carbonica delle *Grid Zones* (Svezia, Germania, Texas).

Per abilitare le chiamate API reali, creare un file `.env` nella directory principale (root) del progetto:
```env
API_KEY=la_tua_chiave_api_qui