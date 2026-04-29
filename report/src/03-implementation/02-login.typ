== Autentificare (Login)

Endpoint-ul `POST /login` este proiectat să reziste tentativelor de compromitere prin
forță brută și analiză de tip side-channel (timing attacks). Implementarea din
`internal/auth/login.go` separă procesul de autentificare în mai multe straturi de
verificare, asigurând un răspuns uniform către utilizator indiferent de natura eșecului.

Pentru a asigura un nivel ridicat de protecție, procesul de autentificare urmează o
secvență strictă de validări, de la rate limiting la nivel de rețea până la hashing-ul
constant-time la nivel de aplicație.

#figure(
  caption: [Diagramă de Secvență - Procesul de Autentificare și Atenuarea Timing Attacks],
  include "assets/02-login-diagram.typ",
) <fig-login-flow>

=== Atenuarea Atacurilor de Timing

O vulnerabilitate comună în sistemele de autentificare este diferența măsurabilă în timpul
de răspuns între un email inexistent și o parolă greșită. Mithras neutralizează acest
vector prin două mecanisme complementare. Primul mecanism, *Dummy Hash*, se activează dacă
utilizatorul nu este găsit în baza de date; în acest caz, sistemul execută
`verifyPassword()` utilizând un hash Argon2id pre-generat cu parametri identici. Astfel,
timpul de procesare rămâne constant la nivel computațional, reducând diferența temporală
la nivel de micro-jitter de rețea și făcând enumerarea utilizatorilor statistic
inaplicabilă pentru un atacator remote. Al doilea mecanism, *Comparație în Timp Constant*,
presupune că rezultatul hash-ului calculat este comparat cu cel stocat folosind
`subtle.ConstantTimeCompare` din biblioteca standard Go. Această funcție asigură că timpul
de execuție al comparației nu depinde de numărul de octeți care coincid, prevenind
atacurile de tip timing care ar putea dezvălui fragmentar conținutul hash-ului.

#figure(
  caption: [Utilizarea `dummyHash` și `subtle.ConstantTimeCompare` în `internal/password/hashed.go`],
)[
  ```go
  func (h Hashed) Verify(raw Raw) (bool, error) {
      target := h.value
      if target == "" {
          target = dummyHash // Execuție Argon2id chiar dacă user-ul nu există
      }

      // ... calcul computed hash ...

      // Comparație rezistentă la timing attacks
      match := subtle.ConstantTimeCompare(dec.hash, computed) == 1
      return match, nil
  }
  ```
]

=== Principiul Fail-Closed

O decizie critică de design în versiunea securizată este aplicarea principiului
*Fail-Closed*. În cazul în care o componentă critică de infrastructură, precum Redis
pentru rate limiting sau PostgreSQL pentru verificarea stării contului, este
indisponibilă, serverul refuză cererea de autentificare (HTTP 500/503). Această abordare
este preferată în locul unei degradări "permisive" (*Fail-Open*), deoarece absența
controalelor de securitate ar lăsa sistemul complet vulnerabil la atacuri de forță brută
în perioadele de instabilitate a infrastructurii.

=== Account Lockout și Rate Limiting

Pentru a bloca atacurile de tip brute-force distribuite, sistemul aplică două bariere
principale. Prima barieră este *Rate Limiting-ul per IP*, implementat la nivel de
middleware, care protejează împotriva inundării serverului cu cereri. A doua barieră este
*Account Lockout-ul*, prin care, după 5 tentative de autentificare eșuate, contul este
blocat automat pentru o perioadă de 15 minute. Această stare este persistentă în baza de
date (coloana `locked_until`), fiind verificată înaintea oricărei procesări de parolă.

=== Răspuns Uniform (Generic Error)

Toate erorile de autentificare sunt returnate sub forma unui cod de eroare generic,
`invalid_credentials` (HTTP 401). Această practică elimină orice oracol de
informații care ar putea fi utilizat pentru *User Enumeration*, asigurând că un atacator
nu poate distinge între diversele cauze de eșec ale autentificării.
