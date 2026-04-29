== Integrare CI/CD cu Analiză Statică de Securitate (SAST)

Garantarea calității codului și identificarea proactivă a vulnerabilităților (*Shift-Left
Security*) au fost asigurate prin integrarea unui pipeline de integrare continuă (CI)
bazat pe *GitHub Actions*.

Pipeline-ul, definit în fișierul `.github/workflows/ci.yaml`, include:
1. *Formatare și Linting*: Analiza de stil folosind `golangci-lint` pentru a asigura
  uniformitatea codebase-ului.
2. *Testare Automată*: Rularea suitelor de teste din Go (`go test ./...`) pentru
  prevenirea regresiilor funcționale.
3. *SAST (Static Application Security Testing)*: Scanarea profundă a surselor cu ajutorul
  utilitarului `gosec`. Acesta detectează automat anti-pattern-uri de securitate precum
  credențiale hardcodate, generatoare de numere pseudo-aleatoare nesigure (ex. `math/rand`
  în contexte criptografice), injecții SQL și referințe slabe la memorie.
4. *Verificarea Dependențelor*: Rularea `govulncheck` pentru a interoga baza de date de
  vulnerabilități cunoscute (CVE-uri) din modulele Go incluse.

Prin această configurare, nicio modificare de cod (*Pull Request* sau *Commit*) nu poate
fi acceptată pe branch-ul `main` dacă introduce o slăbiciune structurală detectabilă
static.

```yaml
jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - uses: ./.github/actions/setup-nix
      - name: GolangCI-lint
        run: nix develop --command golangci-lint run --build-tags=dev ./...

  test-race:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - uses: ./.github/actions/setup-nix
      - name: Run Race Detector
        run: nix develop --command go test -tags=dev -race -v -short ./...
```

Rulează automat mediul de dezvoltare strict și determinist gestionat prin Nix. Fiecare pas
validează că aplicația trece setul complex de reguli analitice și testele de integritate
în concurență (`-race`) înainte să ajungă în branch-ul principal.

#figure(
  caption: [Execuția cu succes a pipeline-ului CI (`ci.yaml`) pe branch-ul `main`: lint, format, test și test-race finalizate integral.],
  image("assets/02-cicd-ci.png"),
)

#figure(
  caption: [Pipeline-ul de audit de securitate (`audit.yaml`): scanare SAST cu `gosec`, verificare dependențe cu `govulncheck` și actualizare Nix.],
  image("assets/02-cicd-sast.png"),
)
