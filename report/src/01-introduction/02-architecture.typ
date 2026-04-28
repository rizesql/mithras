== Arhitectura de Sistem și Modelul de Securitate

Arhitectura sistemului Mithras este fundamentată pe principiul separării
responsabilităților (_Separation of Concerns_) și pe izolarea strictă a limitelor de
încredere (_Trust Boundaries_). Proiectul abandonează modelul monolit în favoarea unei
arhitecturi decuplate, capabile să reziste atacurilor specifice mediilor web moderne prin
utilizarea unor primitive criptografice de ultimă generație și a unor șabloane de design
orientate pe securitate.

#figure(
  caption: [Stack tehnic],
  table(
    columns: (auto, auto, 1fr),
    table.header[*Componentă*][*Versiune*][*Rol*],
    [Go], [1.26.2], [Runtime Identity Provider - logică auth, semnare token],
    [PostgreSQL], [18], [Stocare persistentă: utilizatori, sesiuni, token-uri de resetare],
    [Redis], [8], [Rate limiting - token bucket implementat prin script Lua atomic],
    [ClickHouse / HyperDX], [-], [Telemetrie OpenTelemetry: trace-uri, metrici, log-uri],
    [SvelteKit], [2], [Consumer BFF - gestionare sesiune browser, cookie HttpOnly],
  ),
)

#figure(
  placement: auto,
  scope: "parent",
  caption: [Arhitectura Mithras - limite de încredere și fluxuri de date],
  include "assets/02-architecture-diagram.typ",
)

=== Modelul de Încredere și Suprafața de Atac

Sistemul implementează un model de încredere tripartit, ierarhizat în funcție de
capacitatea mediului de a proteja materialul criptografic sensibil:

1. _Identity Provider (Mithras / Go)_: Reprezintă *Sursa de Încredere* (_Root of Trust_). Este singura entitate autorizată să gestioneze secretele pe termen lung (procesând parolele în clar exclusiv în memoria volatilă, pe durata derivării). Operațiunile critice (generarea cheilor Ed25519, hashing-ul Argon2id și emiterea token-urilor JWS) sunt centralizate și izolate la acest nivel.
2. _Resource Server (BFF / SvelteKit)_: Acționează ca un *Client Confidențial*. Deși nu interacționează niciodată cu parolele utilizatorilor, acesta deține capacitatea de a valida aserțiunile de identitate prin verificarea locală a semnăturilor digitale, utilizând un set de chei publice (JWKS) expus de IdP.
3. _User Agent (Browser)_: Definit formal ca un mediu de execuție nesigur (_Untrusted Environment_). Designul interzice expunerea oricărui material criptografic neopac (precum token-urile Bearer/JWS) către execuția JavaScript, eliminând astfel riscul de furt al identității digitale prin vulnerabilități de tip Cross-Site Scripting (XSS).

=== Șablonul Backend-for-Frontend (BFF)

Integrarea arhitecturii BFF facilitează gestionarea securizată a sesiunilor. BFF-ul acționează
ca un intermediar, stocând token-ul JWS local și expunând browserului exclusiv un cookie
opac `HttpOnly`. Astfel, se mediază tranziția de la starea browserului la serviciile de
backend.

Prin impunerea flag-urilor `HttpOnly`, `Secure` și `SameSite=Strict` asupra cookie-urilor
de sesiune, se creează o segregare logică între codul client și identificatorul de sesiune.
Această decizie neutralizează atacurile de tip _Session Hijacking_ bazate pe extragerea
datelor din `LocalStorage`, atenuând o vulnerabilitate structurală a aplicațiilor Single
Page Application (SPA) convenționale.

=== Protocolul de Autentificare: OAuth 2.0

Mithras implementează OAuth 2.0 Authorization Code Flow: browser-ul este redirecționat
către IdP, care emite un cod de autorizare single-use (valabil 5 minute, stocat cu hash
în baza de date), schimbat de BFF pe un token de acces JWS și un token de refresh.
Mecanismul complet, inclusiv gestionarea stării criptate AES-GCM și fluxul de token
exchange, este detaliat în secțiunea de implementare MVP.

=== Securitatea Aserțiunilor: EdDSA și JWS

Sistemul utilizează formatul JSON Web Signature (JWS) @rfc7515 pentru propagarea
identității între IdP și Resource Server, optând pentru algoritmul asimetric
_EdDSA (Ed25519)_ @rfc8032. Consumer-ul verifică semnăturile local, utilizând cheile
publice expuse de endpoint-ul `/jwks.json`, fără apel back-channel către IdP la
fiecare request.

Payload-ul token-ului respectă principiul minimizării datelor (_Data Minimization_):
sunt expuse exclusiv emitentul (`iss`), identificatorul utilizatorului (`sub`),
marcajele temporale de emitere și expirare (`iat`, `exp`) și aserțiunile de autorizare
(`roles`). Token-ul de acces are o durată de viață de 5 minute. Revocarea sesiunilor
este realizată prin înregistrări persistente în baza de date, indexate după hash-ul
SHA-256 al token-ului de refresh, verificate la fiecare request protejat.
