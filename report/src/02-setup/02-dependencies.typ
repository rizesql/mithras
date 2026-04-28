#import "@preview/codly:1.3.0": *

== Dependențe și Toolchain

=== Toolchain de Dezvoltare

#figure(
  caption: [Unelte de sistem instalate independent de modulul Go],
  table(
    columns: (auto, auto, 1fr),
    table.header[*Unealtă*][*Versiune*][*Rol*],
    [Go], [1.26.2], [Compilator și runtime IdP],
    [Node.js], [24.15.0], [Runtime SvelteKit consumer],
    [pnpm], [10.33.0], [Package manager Node.js],
    [Docker], [28.5.2], [Container runtime],
    [Docker Compose], [v2], [Orchestrare servicii locale],
    [just], [1.46.0], [Task runner; înlocuitor `Makefile`],
    [golangci-lint], [2.11.2], [Linter static Go; instalat prin Nix Flake],
  ),
)

Utilitarele de dezvoltare Go: `goose`, `sqlc`, `gosec`, `govulncheck` și `oapi-codegen`
nu apar în tabel deoarece sunt gestionate direct de toolchain-ul Go prin blocul `tool`
din `go.mod`:

#figure(
  caption: [Bloc `tool` din `go.mod`; utilitare Go gestionate de toolchain],
)[
  ```go
  tool (
      github.com/pressly/goose/v3/cmd/goose
      github.com/sqlc-dev/sqlc/cmd/sqlc
      github.com/securego/gosec/v2/cmd/gosec
      golang.org/x/vuln/cmd/govulncheck
      github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen
  )
  ```
] <go-tool-block>

Versiunile exacte sunt fixate în `go.sum` împreună cu restul dependențelor modulului.
Orice apel `go tool <utilitar>` descarcă și compilează automat versiunea corectă;
nu este necesară instalarea manuală și nu pot apărea divergențe de versiune.

Mediul de dezvoltare este declarat integral și izolat prin *Nix Flake* (`flake.nix`).
Pe mașina virtuală NixOS, utilitarele de sistem din tabel (Go, Node.js, Docker, just, golangci-lint) sunt descărcate și fixate la versiuni exacte prin `flake.lock`, garantând 100%
reproductibilitate fără instalări manuale globale. Activarea mediului se face cu:

#figure(
  caption: [Activare mediu Nix - toolchain izolat la versiuni exacte din `flake.lock`],
)[
  ```sh
  nix develop
  ```
]

=== Generare Cod

Proiectul utilizează generare de cod pentru două componente critice:

*sqlc* - transformă fișierele `.sql` din `pkg/db/migrations/` și interogările din
`pkg/db/queries/` în cod Go cu tipuri stricte. Fiecare parametru SQL devine un câmp
tipizat în Go; concatenarea de string-uri în interogări este prevenită la nivel de generare a codului.

*OpenAPI* - specificația `openapi.yaml` este compilată în tipuri Go pentru
validarea request-urilor la runtime.

Ambele artefacte sunt regenerate cu:

#figure(
  caption: [Regenerare artefacte `sqlc` și `oapi-codegen`],
)[
  ```sh
  just gen
  ```
]

Artefactele generate sunt comise în repository; `just gen` trebuie rulat după
orice modificare a schemei SQL sau a specificației OpenAPI.
