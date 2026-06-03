# Servis za hosting konfiguracija - Alati za Razvoj Softvera 2026

- Učesnici: 
  - Konstantin Kikovski SR-56/2023
  - Jovan Zorić SR-32/2023
  - Kristijan Đeri SR-41/2023

# Opis servisa

Servis služi za kreiranje, čuvanje i prikaz konfiguracija. Konfiguracije su opisane svojim nazivom, verzijom i skupom proizvoljnih parametara. Konfiguracije mogu da se dodaju u konfiguracione grupe, koje isto imaju svoju verziju i pružaju sistem pretrage pomoću skupa proizvoljnih labela dodeljenih konfiguracijama u datoj grupi.

# Tehnologije

Sistem je REST API koji je napisan u Golang 1.25, podaci o konfiguracijama i grupama se čuvaju u HashiCorp Consul KV bazi podataka, servis i baza podataka su kontejnerizovani u Dockeru, a sistem je orkestriran putem docker-compose alata.


# Pokretanje sistema

-> Aktivira se Docker Engine
-> Terminal komanda u root folderu: "docker compose up --build"
-> API servisu se pristupa putem "http://localhost:8000/"
-> Consul bazi podataka se pristupa putem "http://localhost:8500/"
-> Jaeger Tracing servisu se pristupa putem "http://localhost:16686"

# Docker

Sastoji se od tri kontejnera, jedan za Go API servis, drugi za Consul bazu podataka, treći za Jaeger Tracing servis. Go API kontejner je build-ovan u multi-stage, gde prvi stage sadrži celokupan Go jezik, compiler, biblioteke i alate, a drugi stage samo pokreće binary executable file koji Go compiler pravi, stoga koristi mnogo manje resursa. Sistem je orkestriran putem docker-compose alata, gde se Consul kontejner pokreće prvi, Jaeger kontejner drugi, a Go API kontejner treći, i tako uspostavljaju mrežu i redosled zavisnosti.

# API endpoints (http://localhost:8000/)

- Konfiguracije:

  - GET all configs -> dobavlja sve konfiguracije u sistemu
    - Request: GET "/configs"
    - Response 200: vraća JSON sa svim konfiguracijama
    - Response 500: vraća grešku da server ne odgovara

  - GET config by name -> vraća sve verzije jedne konfiguracije
    - Request: GET "/configs/{configName}"
    - Response 200: vraća JSON sa svim verzijama konfiguracije pod nazivom {configName}
    - Response 404: vraća grešku, konfiguracije nisu pronađene

  - GET config by name and version -> vraća konfiguraciju sa tim {configName} i tim {configVersion}
    - Request: GET "/configs/{configName}/{configVersion}"
    - Response 200: vraća JSON sa konfiguracijom pod nazivom {configName} i verzijom {configVersion}
    - Response 404: vraća grešku, konfiguracija nije pronađena

  - POST add config -> kreira konfiguraciju sa zadatim nazivom i verzijom
    - Request: POST "/configs/{configName}/{configVersion}"
      - Body: očekuje proizvoljan broj parametara
        - {
            params: {
              "key1": "value1"
              "key2": "value2"
            }
          }
    - Response 201: vraća potvrdu da je konfiguracija kreirana
    - Response 409: vraća poruku da konfiguracija već postoji pod tim nazivom i verzijom

  - PUT edit config -> menja konfiguraciju pod tim nazivom i verzijom
    - Request: POST "/configs/{configName}/{configVersion}"
      - Body: očekuje celu novu konfiguraciju sa izmenama
    - Response 200: vraća potvrdu da je izmena uspešna
    - Response 404: vraća grešku, konfiguracija nije pronađena

  - DELETE config -> briše konfiguraciju pod tim nazivom i verzijom
    - Request: DELETE "/configs/{configName}/{configVersion}"
    - Response 204: vraća potvrdu da je konfiguracija obrisana
    - Response 404: vraća grešku, konfiguracija nije pronađena

- Konfiguracione grupe:

  - GET all groups -> dobavlja sve grupe u sistemu sa njihovim konfiguracijama
    - Request: GET "/configsGroup"
    - Response 200: vraća JSON sa svim grupama i svim konfiguracijama u tim grupama
    - Response 5xx: vraća grešku da server ne odgovara

  - GET group by name and version -> vraća grupu pod tim nazivom i tom verzijom i konfiguracije u toj grupi
    - Request: GET "/configsGroup/{groupName}/{groupVersion}"
    - Response 200: vraća JSON sa grupom pod tim nazivom i tom verzijom i listom konfiguracija
    - Response 404: vraća grešku, grupa nije pronađena

  - POST add one group -> kreira novu grupu sa tim nazivom i tom verzijom
    - Request: POST "/configsGroup/{groupName}/{groupVersion}"
    - Response 201: vraća potvrdu da je grupa kreirana
    - Response 409: vraća poruku da grupa već postoji pod tim nazivom i verzijom

  - PUT add config to group -> dodaje postojeću konfiguraciju sa tim nazivom i verzijom u grupu sa svojim nazivom i verzijom
    - Request: PUT "/configsGroup/{groupName}/{groupVersion}"
      - Body: očekuje JSON sa postojećom konfiguracijom i JSON listom labela
        - {
            "name": "naziv",
            "version": integer,
            "labels": {
              "label1": "value1",
              "label2": "value2"
            }
        }
    - Response 200: vraća potvrdu da je dodavanje uspešno
    - Response 404: vraća grešku, grupa nije pronađena ili konfiguracija nije pronađena
  
  - GET all configs in group by labels -> dobavlja sve konfiguracije u datoj grupi prema navedenim labelama
    - Request: GET "/configsGroup/{groupName}/{groupVersion}/label1=value1|label2=value2"
    - Response 200: vraća JSON sa svim konfiguracijama u toj grupi koje imaju tačno sve navedene labele
    - Response 404: vraća grešku, grupa nije pronađena

  - DELETE remove config from group by labels -> briše sve konfiguracije u datoj grupi prema navedenim labelama
    - Request: DELETE "/configsGroup/{groupName}/{groupVersion}/label1=value1|label2=value2"
    - Response 204: vraća potvrdu da su konfiguracije obrisane
    - Response 404: vraća grešku, grupa nije pronađena

  - DELETE one group by name and version -> briše grupu pod tim nazivom i verzijom
    - Request: DELETE "/configsGroup/{groupName}/{groupVersion}"
    - Response 204: vraća potvrdu da je grupa obrisana
    - Response 404: vraća grešku, grupa nije pronađena

- Metrike:

  - GET metrics -> dobavlja podatke o metrikama servisa
    - Request: GET "/metrics"
    - Response 200: vraća JSON sa podacima o praćenim metrikama
    - Response 500: vraća grešku da server ne odgovara