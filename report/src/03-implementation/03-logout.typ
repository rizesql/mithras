== Logout

Mecanismul de logout în Mithras este proiectat sub paradigma *Security First*, punând
accent pe invalidarea completă a stării sesiunii atât în baza de date, cât și la
nivelul întregului cont în cazul detectării unor anomalii. Procedura de logout
implică o coordonare între Consumer (BFF) și Identity Provider (IdP).

=== Procesul de Logout Dual

Deconectarea unui utilizator se realizează prin două acțiuni complementare, fiecare
vizând un strat diferit al arhitecturii. La nivelul *Consumer-ului (BFF)*, logout-ul
presupune ștergerea imediată a cookie-ului `HttpOnly` care stochează token-ul de
sesiune în browserul utilizatorului. Această acțiune termină sesiunea din perspectiva
clientului, făcând accesul imposibil prin interfața web.

La nivelul *Mithras IdP*, logout-ul implică o procedură de revocare server-side a
token-ului de refresh. În versiunea securizată (*main*), acest proces are un efect
de *Force Logout All*, prin care toate sesiunile active ale utilizatorului respectiv
sunt revocate simultan în baza de date prin setarea câmpului `revoked_at` la momentul
curent. Această decizie de design asigură că orice altă sesiune potențial uitată sau
compromisă pe alte dispozitive este terminată imediat, eliminând suprafața de atac
reziduală.

#figure(
  caption: [Diagramă de Secvență - Procedura de Logout și Revocare],
  include "assets/03-logout-diagram.typ",
) <fig-logout-flow>

=== Protecția Token-ului de Refresh: Hashing SHA-256

Mithras aplică principiul stocării sigure nu doar pentru parole, ci și pentru
token-urile de refresh. Serverul nu stochează niciodată valoarea brută a token-ului
în baza de date. În schimb, se calculează și se salvează hash-ul *SHA-256* al
acestuia. Această măsură garantează că, în eventualitatea unei scurgeri de date din
PostgreSQL, un atacator nu poate utiliza valorile din tabelul `sessions` pentru a
efectua operațiuni de logout sau pentru a genera noi token-uri de acces.

=== Rate Limiting

Endpoint-ul de logout este protejat prin *Rate Limiting* pentru a preveni abuzurile și
epuizarea resurselor prin cereri de revocare repetate.

=== Compromisul Stateless: Token-uri de Acces JWS

Mithras utilizează token-uri de acces JWS semnate asimetric (EdDSA), care sunt prin
definiție *stateless*. Acest lucru permite Consumer-ului să verifice validitatea
unui token local, utilizând cheile publice ale IdP-ului, fără a efectua un apel de
rețea la fiecare cerere, ceea ce îmbunătățește semnificativ performanța și
scalabilitatea.

Totuși, natura stateless introduce un compromis de securitate: un token de acces odată emis
rămâne valid până la expirare. Pentru a mitiga acest risc, Mithras impune o durată de
viață scurtă (5 minute). Astfel, deși revocarea accesului la nivelul IdP-ului este imediată
prin invalidarea token-ului de refresh, accesul efectiv este terminat doar după expirarea
token-ului de acces curent.

=== Detecția Anomaliilor la Logout

Dacă un token de refresh este prezentat la logout, dar acesta apare deja ca fiind
revocat în baza de date, sistemul declanșează o procedură de protecție împotriva
*Token Replay*. Reutilizarea unui token de refresh indică o tentativă de fraudă sau
un furt de sesiune. În acest scenariu, Mithras invalidează preventiv toate sesiunile
active ale utilizatorului și refuză cererea, marcând incidentul în telemetrie pentru
investigații ulterioare.

#figure(
  caption: [Implementare Logout cu invalidare totală în `internal/auth/logout.go`],
)[
  ```go
  func (l *Logout) Logout(ctx context.Context, rawToken string) (err error) {
      // ... identificare sesiune după hash ...

      // Revocare preventivă a TUTUROR sesiunilor utilizatorului
      revErr := db.Query.RevokeUserSessions(ctx, l.db, db.RevokeUserSessionsParams{
          Now:    &now,
          UserPk: sess.UserPk,
      })

      return nil
  }
  ```
]
