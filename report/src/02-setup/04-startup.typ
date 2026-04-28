== Strategia de Branching și Izolarea Versiunilor

Proiectul menține două versiuni distincte ale codului sursă pe branch-uri Git separate.
Această separare este cerința fundamentală a metodologiei Build-Hack-Secure: dovezile
de atac și cele de remediere trebuie să fie reproductibile independent, pe cod identificabil.

#figure(
  caption: [Branch-uri Git; caracteristici de securitate per versiune],
  table(
    columns: (auto, auto, 1fr),
    table.header[*Branch*][*Rol*][*Caracteristici cheie*],
    [`main`],
    [V2 Securizat],
    [Argon2id, rate limiting Redis, mesaj generic la login, token reset criptografic, cookie HttpOnly/Secure/SameSite=Strict.],

    [`vulnerable`],
    [V1 Deliberat vulnerabil],
    [Stocare parolă în clar, NoopStore, mesaje de eroare diferențiate, token reset predictibil, cookie fără flag-uri de securitate.],
  ),
)

Comutarea între versiuni se face exclusiv prin:

#figure(
  caption: [Comutare între versiunile V1 și V2],
)[
  ```sh
  # Activare versiune vulnerabilă (PoC-uri)
  git checkout vulnerable

  # Revenire la versiunea securizată (re-test)
  git checkout main
  ```
]

Baza de date nu este resetată automat la comutarea branch-ului. Între rularea PoC-urilor
pe V1 și re-testarea pe V2 este necesară curățarea datelor pentru a elimina starea
reziduală (conturi create cu parole triviale, sesiuni active, token-uri de resetare
neutilizate):

#figure(
  caption: [Resetare completă a mediului între sesiunile de testare],
)[
  ```sh
  just nuke   # docker compose down -v — șterge volumele
  just up     # repornire cu schemă curată
  go run ./cmd/mithras datastore migrate
  ```
]

=== Modificări deliberate pe branch-ul `vulnerable`

Tabelul următor listează fiecare vulnerabilitate, fișierul modificat și natura exactă
a schimbării față de `main`. Nicio modificare din această listă nu reprezintă o
simplificare accidentală, fiecare este intenționată și documentată.

#figure(
  caption: [Vulnerabilități injectate pe branch-ul `vulnerable`, fișier modificat și natura schimbării],
  table(
    columns: (auto, auto, 1fr),
    table.header[*ID*][*Fișier modificat*][*Modificare față de `main`*],

    [VULN-01],
    [`internal/password/raw.go`],
    [Funcția `New()` acceptă orice string non-gol. Validarea lungimii și complexității este eliminată.],

    [VULN-02],
    [`internal/password/hashed.go`],
    [Funcția `Hash()` returnează valoarea raw a parolei fără transformare. Câmpul `secret` din DB conține parola în clar.],

    [VULN-03],
    [`internal/ratelimit/`],
    [`Policy.Store` este înlocuit cu un `NoopStore` care aprobă orice request fără contorizare. Middleware-ul rămâne structural prezent.],

    [VULN-04],
    [`internal/auth/login.go`],
    [`fetchUser()` returnează `errUserNotFound` distinct față de `errWrongPassword`. Verificarea parolei pe dummy hash este eliminată, timing diferit măsurabil.],

    [VULN-05],
    [`client/src/` (consumer)],
    [Cookie setat fără `HttpOnly`, `Secure` sau `SameSite`. Token de acces cu expirare extinsă la 24h. Logout nu revocă sesiunile.],

    [VULN-06],
    [`internal/auth/password_reset.go`],
    [Token generat ca `base64url(email + unix_timestamp)`. Nu expiră, nu este marcat `used_at`, nu invalidează sesiunile la reset.],
  ),
)

=== Verificarea branch-ului activ

Înainte de orice captură de ecran, branch-ul activ este confirmat explicit:

#figure(
  caption: [Confirmare branch activ înaintea oricărei capturi de ecran],
)[
  ```sh
  git branch --show-current
  # expected: vulnerable  (pentru PoC-uri)
  # expected: main        (pentru re-test)
  ```
]

Toate capturile din secțiunile de PoC au fost realizate cu output-ul acestei
comenzi vizibil în terminal, confirmând că atacul rulează pe versiunea vulnerabilă.
