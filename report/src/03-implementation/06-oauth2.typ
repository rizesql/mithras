== Protocolul OAuth 2.0 și PKCE

Mithras nu este doar un serviciu de autentificare izolat, ci un furnizor de identitate
(*Identity Provider*) conform cu standardul *OAuth 2.0* @rfc6749. Implementarea vizează
securizarea fluxului de delegare a accesului către aplicații terțe (precum Consumer-ul BFF).

=== Authorization Code Flow cu PKCE

Sistemul utilizează fluxul de tip *Authorization Code*. O particularitate a implementării
Mithras este impunerea obligatorie a extensiei *PKCE* (Proof Key for Code Exchange)
@rfc7636, chiar și pentru clienții confidențiali.

PKCE adaugă un strat de verificare criptografică prin utilizarea unui `code_challenge`
(transmis în faza de autorizare) și a unui `code_verifier` (prezentat în faza de schimb de
token). Această metodă garantează că doar entitatea care a inițiat fluxul de autorizare
poate finaliza schimbul de token-uri, neutralizând atacurile de tip *Authorization Code
Injection*.

=== Gestiunea Stării Criptate (AES-GCM)

Pentru a menține integritatea fluxului de autorizare între redirecționări, Mithras
utilizează un mecanism de *Encrypted State*. Atunci când un utilizator accesează
`/authorize`, parametrii cererii (Client ID, Redirect URI, PKCE Challenge) sunt
împachetați într-un obiect de stare, serializați și criptați utilizând algoritmul
*AES-256-GCM*. Utilizarea modului GCM asigură criptare de tip *AEAD* (Authenticated
Encryption with Associated Data), adăugând nu doar confidențialitate datelor, ci și un tag
de autentificare strict.

Valoarea criptată este stocată într-un cookie securizat (`Auth-State`), asigurând
următoarele proprietăți de securitate:
1. *Confidențialitate*: Detaliile fluxului nu sunt vizibile pentru utilizator sau
  atacatori externi.
2. *Integritate*: Orice modificare a stării în timpul redirecționării (ex. atacuri de tip
  Bit-Flipping sau Padding Oracle) este detectată prin invalidarea tag-ului GCM,
  ducând la eșuarea decriptării și respingerea automată a cererii.
3. *Legarea Sesiunii*: Cookie-ul este marcat `HttpOnly` și `Secure`, prevenind
  accesul din JavaScript și limitând expunerea la medii criptate.

=== Securitatea Aserțiunilor: JWS EdDSA

Rezultatul final al fluxului OAuth2 este emiterea unui token de acces în format JWS.
Mithras optează pentru algoritmul asimetric *EdDSA* (Ed25519) @rfc8032, care oferă
avantaje semnificative față de RS256 sau HS256:
- *Performanță*: Semnăturile EdDSA sunt mai rapide și utilizează chei mult mai scurte
  pentru același nivel de securitate.
- *Verificare Locală*: Aplicațiile consumatoare pot valida token-urile utilizând exclusiv
  cheia publică a IdP-ului (expusă prin `/jwks.json`), eliminând necesitatea unui apel
  back-channel către Mithras la fiecare request protejat.
- *Rezistență*: Spre deosebire de HS256 (simetric), compromiterea unui Resource Server nu
  periclitează capacitatea de a emite token-uri noi, deoarece cheia privată rămâne izolată
  în memoria IdP-ului.

#figure(
  caption: [Implementarea PKCE Challenge în `internal/auth/oauth2.go`],
)[
  ```go
  func (o *OAuth2) MintCode(ctx context.Context, userPk int64, state AuthorizeState) (CodeID, error) {
      code := idkit.NewAuthorizationCodeID()
      // Stocare atomică a codului legat de challenge-ul PKCE
      _, err := db.Query.InsertAuthorizationCode(ctx, o.db, db.InsertAuthorizationCodeParams{
          Code:      code,
          UserPk:    userPk,
          Challenge: state.Challenge, // PKCE Challenge
          ExpiresAt: time.Now().Add(5 * time.Minute),
      })
      return code, err
  }
  ```
]
