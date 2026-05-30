# Servis za hosting konfiguracija - Alati za Razvoj Softvera 2026

- Učesnici: 
  - Konstantin Kikovski SR-56/2023
  - Jovan Zorić SR-32/2023
  - Kristijan Đeri SR-41/2023

# Opis servisa

Servis služi za kreiranje, čuvanje i prikaz konfiguracija. Konfiguracije su opisane svojim nazivom, verzijom i skupom proizvoljnih parametara. Konfiguracije mogu da se dodaju u konfiguracione grupe, koje isto imaju svoju verziju i pružaju sistem pretrage pomoću skupa proizvoljnih labela dodeljenih konfiguracijama u datoj grupi.

# Tehnologije

Sistem je REST API koji je napisan u Golang 1.25, podaci o konfiguracijama i grupama se čuvaju u HashiCorp Consul KV bazi podataka, servis i baza podataka su kontejnerizovani u Dockeru, a sistem je orkestriran putem docker-compose alata.

# Arhitektura sistema

// TODO

# Pokretanje sistema

-> Aktivira se Docker Engine
-> Terminal komanda u root folderu: "docker compose up --build"
-> API servisu se pristupa putem "http://localhost:8000/"
-> Consul bazi podataka se pristupa putem "http://localhost:8500/"

# API endpoints (http://localhost:8000/)

- Konfiguracije:

- GET "/configs" -> dobavlja sve konfiguracije u sistemu
- GET "/configs/{name}/{version}" -> dobavlja jednu konfiguraciju, po nazivu i verziji
- POST "/configs" -> kreira novu konfiguraciju
- PUT "/configs/{name}/{version}" -> menja postojeću konfiguraciju, po nazivu i verziji (imutabilno, celokupna zamena)
- DELETE "/configs/{name}/{version}" -> briše postojeću konfiguraciju, po nazivu i verziji

- Konfiguracione grupe:

- TODO

# Request forme primeri

- POST "/configs" body:

// TODO

# Response forme primeri

// TODO

# Response code značenja

200: uspešno dobavljeni podaci
201: uspešno kreiran entitet
204: uspešno obrisan entitet

404: nisu pronađeni podaci
429: preopterećenje sistema brojem zahteva u sekundi

5xx: interna greška u serveru

# Docker

Sastoji se od dva kontejnera, jedan za Go API servis, drugi za Consul bazu podataka. Go API kontejner je build-ovan u multi-stage, gde prvi stage sadrži celokupan Go jezik, compiler, biblioteke i alate, a drugi stage samo pokreće binary executable file koji Go compiler pravi, stoga koristi mnogo manje resursa. Sistem je orkestriran putem docker-compose alata, gde se Consul kontejner pokreće prvi, a Go API kontejner drugi, i komuniciraju u mreži 

// TODO