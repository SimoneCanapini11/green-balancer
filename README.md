# Carbon-Aware Load Balancer

Questo repository contiene l'implementazione di un Reverse Proxy sviluppato in **Go**, progettato per integrare i principi del *Green Cloud Computing* nel routing di rete. 

Il sistema è dotato di logiche di instradamento dinamico *Carbon-Aware*, bilanciando il traffico verso le regioni Cloud con la minore intensità carbonica in tempo reale. Per garantire High Availability e prevenire fallimenti a cascata, l'architettura implementa pattern di resilienza avanzati, tra cui *Circuit Breaker* e *Graceful Degradation*. L'infrastruttura è stata collaudata sia in ambiente locale tramite container, sia in produzione su istanze Amazon Web Services (AWS EC2).

---

## 1. Prerequisiti

Per eseguire l'architettura in locale e lanciare le suite di test, è necessario aver installato sulla propria macchina:
- **Docker** e **Docker Compose** (per l'orchestrazione dei container)
- **Go** (versione 1.20 o superiore, per eseguire gli Unit Test isolati)
- **k6** (framework open-source di Grafana Labs, per gli Stress Test)

---

## 2. Configurazione Iniziale

Il Load Balancer interroga in tempo reale l'API (Electricity Maps) per ottenere l'intensità carbonica delle *Grid Zones* (Svezia, Germania, Texas).

Per abilitare le chiamate API reali, copiare il file `.env` nella directory principale (root) del progetto, sostituendolo al file `.env.example`:
```env
API_KEY=la_chiave_api_va_qui
```
**Nota sulla sicurezza:** Il file `.env` è stato inserito nel `.gitignore` e non verrà tracciato dal controllo di versione.

**Nota di Resilienza:** L'inserimento della chiave non è strettamente obbligatorio per testare il routing. In assenza di una chiave valida (o in caso di API offline), il Load Balancer attiverà in automatico i pattern di Fallback, basandosi su valori emissivi storici simulati, garantendo il continuo instradamento del traffico senza errori a runtime.

---

## 3. Start in ambiente locale

L'intera architettura è riproducibile localmente con un singolo comando.

Aprire il terminale nella root del progetto ed eseguire:
```bash
docker-compose up --build -d
```

Il Load Balancer sarà in ascolto sulla porta configurata (es. `localhost:8080`).

Per arrestare l'ambiente e pulire le risorse allocate (network e container):
```bash
docker-compose down
```

---

## 4. Cloud Deployment (Produzione su AWS EC2)
Per replicare l'ambiente di produzione utilizzato durante la fase di valutazione architetturale su Amazon Web Services, seguire questi passaggi operativi:

**1. Provisioning dell'Istanza:** Avviare un'istanza Amazon EC2 (es. t3.micro) con Ubuntu Server.

**2. Security Groups:** Assicurarsi che il Security Group associato all'istanza consenta il traffico in entrata sulla porta 80/8080 (HTTP) e sulla porta 22 (SSH).

**3. Accesso e Setup dell'Host:** Collegarsi all'istanza tramite la console AWS. Una volta all'interno del terminale, installare le dipendenze necessarie:
```bash
# Aggiornare il sistema
sudo apt update && sudo apt upgrade -y
   
# Installare Docker, Docker Compose e Git
sudo apt install docker.io docker-compose-v2 git -y
```

**4. Deploy del Codice:** Clonare questo repository sull'istanza e avviare l'infrastruttura in modalità detached:
```bash
git clone https://github.com/SimoneCanapini11/green-balancer.git
cd green-balancer
sudo docker compose up --build -d
```

Il Reverse Proxy Carbon-Aware sarà raggiungibile pubblicamente tramite l'IP pubblico dell'istanza EC2.

---

## 5. Testing e Validazione
**Validazione della Resilienza (Fault Injection)**
Per testare il comportamento del sistema in condizioni di guasto (attivazione del Circuit Breaker, mantenimento della latenza azzerata tramite Stale Cache e Fallback statistico), è stata sviluppata una suite di API Mocking.
Posizionarsi nella root del progetto ed eseguire:
```bash
go test ./strategy -v (test per carbonaware_test.go)

go test ./emissions -v (test per emissions_test.go)

go test ./balancer -v (test per integration_test.go)
```

**Stress Test e Validazione dello Spillover (Chaos Engineering)**
Per verificare la corretta attivazione della logica di Spillover (dirottamento del traffico verso il secondo nodo più ecologico in caso di saturazione delle connessioni concorrenti sul nodo primario), utilizzare il framework k6.
Assicurandosi che i container siano in esecuzione (docker-compose up -d), lanciare lo script di carico su un secondo terminale:
```bash
k6 run test/load_test.js 
```
I risultati evidenzieranno la perfetta distribuzione dinamica delle richieste in base al limite di tolleranza di ciascun worker.
