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

Il Load Balancer interroga in tempo reale l'API (Electricity Maps) per ottenere l'intensità carbonica delle *Grid Zones* (Svezia, Germania, Texas).

Per abilitare le chiamate API reali, creare un file `.env` nella directory principale (root) del progetto:
```env
API_KEY=la_tua_chiave_api_qui
```

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

Il Reverse Proxy Carbon-Aware sarà ora operativo e raggiungibile pubblicamente tramite l'IP pubblico (o il record DNS) dell'istanza EC2.

