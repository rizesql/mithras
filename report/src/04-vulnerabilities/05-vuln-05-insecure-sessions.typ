== VULN-05: Gestionare Nesigură a Sesiunilor (Insecure Session Management)

Sistemul Consumer-ului stochează token-ul de acces utilizând un cookie care nu posedă
flag-urile de securitate esențiale (`HttpOnly`, `Secure`, `SameSite`). Pe latura server-
ului IdP, token-ul de acces este configurat pe versiunea V1 cu o durată de viață extinsă
de 24 de ore, iar operațiunea de logout nu execută nicio invalidare a sesiunilor în baza
de date. Defectul este clasificat drept A07:2025 și CWE-614 (Sensitive Cookie in HTTPS
Session Without Secure Attribute) @cwe614.

=== Fundamentare Teoretică și Implementare Defectuoasă

Gestiunea stării pe web necesită protejarea identificatorilor de sesiune împotriva
scurgerilor de informații prin execuția de cod malițios sau interceptare de trafic.
Flag-ul `HttpOnly` interzice motorului JavaScript din browser să citească valoarea
cookie-ului, fiind un strat vital de atenuare a impactului atacurilor de tip Cross-Site
Scripting (XSS).

În implementarea vulnerabilă, setarea cookie-ului este superficială:

#figure(
  caption: [Implementare V1 - Setarea cookie-ului fără flag-uri de securitate],
)[
  ```javascript
  // client/src/lib/auth.server.ts (branch: vulnerable)
  cookies.set('access_token', token, {
      path: '/',
      // HttpOnly: false, -> permite accesul din document.cookie
      // Secure: false, -> permite transmiterea pe HTTP necriptat
      // SameSite: 'Lax', -> vulnerabil la CSRF limitat
      maxAge: 60 * 60 * 24 // Expirare extinsă nejustificat (24 de ore)
  });
  ```
]

La nivel de backend, revocarea nu este respectată:

#figure(
  caption: [Implementare V1 - Logout ineficient fără revocare în baza de date],
)[
  ```go
  // internal/auth/logout.go (branch: vulnerable)
  func (l *Logout) Logout(ctx context.Context, token string) error {
      // Nu marchează token-ul de refresh ca `revoked_at`
      // Nu invalidează sesiunile active din baza de date
      return nil
  }
  ```
]

=== Mecanism de Atac și Exploatare

Absența flag-ului `HttpOnly` expune direct materialul de sesiune către API-ul
`document.cookie`. În cazul în care aplicația client conține o vulnerabilitate secundară
de tip XSS, un atacator poate injecta un payload JavaScript care exportă silențios
token-ul către un server extern. Datorită duratei de viață de 24 de ore a token-ului de
acces JWS și lipsei de invalidare reală la logout, atacatorul poate utiliza token-ul furat
pentru a accesa contul victimei, independent de acțiunile acesteia de a se deconecta din
aplicație.
