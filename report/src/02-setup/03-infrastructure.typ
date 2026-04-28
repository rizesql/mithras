== Infrastructură Docker Compose

Serviciile de suport ale aplicației sunt gestionate prin Docker Compose, eliminând
dependințele de sistem și garantând reproductibilitatea mediului. Fișierul
`docker-compose.yml` definește patru servicii, fiecare cu restricții explicite de
resurse și politici `no-new-privileges`.

#figure(
  caption: [Servicii Docker Compose],
  table(
    columns: (auto, auto, 1fr),
    table.header[*Serviciu*][*Port local*][*Rol și configurație notabilă*],
    [`postgres`],
    [5432],
    [PostgreSQL `shared_buffers=256MB, work_mem=16MB`; logging pentru DDL, interogări lente (>1000ms) și așteptări pe lock.],

    [`redis`],
    [6379],
    [Redis. Persistență `appendonly`; limită `maxmemory 256MB` cu politică de evicție `allkeys-lru`.],

    [`clickstack`],
    [8081 / 4317],
    [ClickHouse + HyperDX UI. Receptor OTLP gRPC (4317) și HTTP (4318) pentru trace-uri, metrici și log-uri.],

    [`drizzle`],
    [4983],
    [Drizzle Gateway - interfață vizuală pentru inspecția schemei PostgreSQL în timpul dezvoltării.],
  ),
)

=== Pornirea infrastructurii

Toate serviciile sunt pornite cu o singură comandă:

#figure(
  caption: [Pornire servicii Docker Compose],
)[
  ```sh
  just up
  # echivalent: docker compose up -d
  ```
]

Comanda returnează după ce toate containerele raportează starea `healthy`
(verificată prin `healthcheck` definit per serviciu în `docker-compose.yml`).

#figure(
  caption: [Output `docker compose ps`; toate serviciile în starea `healthy`],
  image("assets/03-infrastructure-docker-ps.png"),
)

=== Migrații bază de date

Schema PostgreSQL este gestionată prin `goose` v3. Migrațiile sunt aplicate
explicit înaintea primei porniri a serverului:

#figure(
  caption: [Aplicare migrații schemă PostgreSQL],
)[
  ```sh
  go run ./cmd/mithras datastore migrate
  ```
]

La pornire, serverul verifică integritatea schemei in baza de date și refuză să
pornească dacă există migrații neaplicate, prevenind incompatibilități între cod și schemă.

#figure(
  caption: [Verificare stare migrații; toate înregistrările trebuie să aibă starea `Applied`],
)[
  ```sh
  go run ./cmd/mithras datastore status
  # afișează versiunea curentă a schemei și starea fiecărei migrații
  ```
]

#figure(
  caption: [Output `datastore status`; toate migrațiile cu starea `Applied`],
  image("assets/03-infrastructure-db-status.png"),
)

=== Pornirea serverului

#figure(
  caption: [Pornire server Mithras; direct sau prin `just run`],
)[
  ```sh
  go run ./cmd/mithras serve
  # sau, în dezvoltare:
  just run   # up + generare cod sqlc/OpenAPI + serve
  ```
]

Serverul citește configurația din `mithras.yaml`:

#figure(
  caption: [Parametri de configurare relevanți `mithras.yaml`],
  table(
    columns: (auto, 1fr),
    table.header[*Parametru*][*Valoare / Semnificație*],
    [`http_port`], [8080: portul de ascultare HTTP],
    [`issuer`], [`http://localhost:8080`: valoarea câmpului `iss` din token-uri JWS],
    [`auth.kek`], [Cheie AES-256 (Base64) pentru criptarea stării OAuth2],
    [`db.uri`], [URI PostgreSQL cu credențiale, schema și tabel de migrații],
    [`ratelimit.type`], [`redis`: backend token bucket pentru rate limiting],
    [`tracing.exporter`], [`otlp`: export trace-uri către ClickHouse prin OTLP gRPC],
  ),
)

La pornire reușită, serverul loghează confirmarea ascultării pe portul 8080 și
conexiunea activă la PostgreSQL, Redis și colectorul OTLP.

#figure(
  caption: [Log-uri de pornire Mithras; confirmare port 8080 și conexiuni active],
  image("assets/03-infrastructure-startup.jpeg"),
)

=== Verificarea stării serviciilor

#figure(
  caption: [Secvență de verificare completă a stack-ului],
)[
  ```sh
  # Starea containerelor
  docker compose ps

  # Starea schemei DB
  go run ./cmd/mithras datastore status

  # Confirmare că serverul răspunde
  curl -s http://localhost:8080/health/live | jq
  ```
]

Serverul expune două endpoint-uri de health:

- `GET /health/live` - *liveness probe*: returnează `200 OK` dacă runtime-ul a
  pornit, `503` dacă se află în starea `idle`. Nu execută verificări externe.
- `GET /health/ready` - *readiness probe*: rulează toate verificările de readiness
  înregistrate (conexiune DB, etc.) și returnează `200` doar dacă toate trec.

Pentru verificarea pornirii se folosește `/health/live` - semantica sa este exactă:
confirmă exclusiv că procesul rulează, fără a presupune starea dependențelor.

#figure(
  caption: [Răspuns `GET /health/live` - HTTP 200, server operațional],
  image("assets/03-infrastructure-health.png"),
)
