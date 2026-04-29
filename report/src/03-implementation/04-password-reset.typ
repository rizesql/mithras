== Resetare Parolă

Procesul de resetare a parolei reprezintă un vector critic de atac în orice sistem de
gestiune a identității, fiind adesea veriga slabă exploatată pentru preluarea conturilor
(*Account Takeover*). Mithras implementează acest flux urmând recomandările
*OWASP Forgot Password Cheat Sheet*, asigurând un echilibru între securitatea
criptografică și protecția împotriva scurgerilor de informații.

=== Fluxul de Resetare și Validare

Arhitectura procesului de resetare este împărțită în două faze distincte: solicitarea
link-ului de resetare și consumul efectiv al token-ului. Coordonarea între actorii
sistemului este ilustrată în diagrama de mai jos:

#figure(
  caption: [Diagramă de Secvență - Fluxul de Resetare Parolă și Revocare],
  include "assets/04-reset-diagram.typ",
) <fig-reset-flow>

=== Generarea și Stocarea Securizată a Token-ului

Atunci când un utilizator solicită resetarea parolei prin endpoint-ul
`POST /forgot-password`, sistemul generează un token format din 32 de octeți de date
aleatorii generate de un CSPRNG. Această decizie asigură o entropie de *256 de biți*,
depășind semnificativ recomandările OWASP de minimum 128 de biți pentru token-uri de
recuperare. Mithras utilizează un format compozit pentru token-ul transmis utilizatorului,
structurat sub forma `{ID}.{Secret}`. Identificatorul public (`ID`) este un *NanoID* (de
exemplu, cu prefixul `rst_`), un format opac, url-friendly și rezistent la atacuri de tip
enumerare.

Această separare între identificator și materialul criptografic secret permite localizarea
rapidă a înregistrării în baza de date fără a expune materialul sensibil în log-uri de
acces sau la nivelul stratului de stocare. În timp ce identificatorul NanoID este utilizat
pentru regăsirea datelor, partea de `Secret` este utilizată exclusiv pentru verificarea
integrității cererii, fiind stocată sub formă de hash.

Pentru a respecta principiul *Defense in Depth*, secretul nu este stocat niciodată în
clar. Mithras aplică o transformare unidirecțională *SHA-256* asupra părții secrete
înainte de persistență. Astfel, chiar și în cazul unei scurgeri de date, atacatorul nu
poate reconstitui materialul necesar pentru a finaliza o resetare neautorizată.

=== Prevenirea Enumerării

O măsură critică de securitate implementată la nivelul endpoint-ului de solicitare este
returnarea unui răspuns uniform de succes indiferent dacă adresa de email există sau nu în
sistem. Această tactică elimină capacitatea unui atacator de a utiliza formularul de
resetare pentru a confirma existența conturilor legitime în baza de date.

=== Validarea și Constrângerile de Securitate

Procesul de resetare finalizat prin `POST /reset-password` impune o serie de constrângeri
suplimentare. Fiecare token este de unică folosință, marcat ca utilizat atomic după prima
resetare reușită, și are o durată de viață limitată la o oră. În momentul resetării,
sistemul invalidează automat orice alte cereri de resetare active emise anterior pentru
același utilizator.

O barieră suplimentară este verificarea istoricului parolelor, unde noua parolă este
supusă derivării criptografice și comparată cu ultimele cinci hash-uri din tabelul
`password_history`. Dacă se detectează o potrivire, cererea este respinsă, forțând
utilizatorul să adopte o identitate digitală nouă și prevenind vulnerabilitățile legate de
rotația între parole cunoscute.

=== Revocarea Sesiunilor Active la Resetare

Finalizarea resetării invalidează instantaneu toate sesiunile active ale utilizatorului,
garantând că orice potențial atacator care deține un token de sesiune valid trebuie să se
reautentifice. Această măsură asigură controlul exclusiv al proprietarului asupra contului.

#figure(
  caption: [Implementarea logică a resetării și revocării în `internal/auth/password_reset.go`],
)[
  ```go
  func (r PasswordReset) performReset(..., rst db.PasswordResetRow, newSecret *password.Hashed) error {
      return db.Tx(ctx, r.db, func(tx db.DBTX) error {
          // Actualizare credențial și marcare token ca utilizat
          db.Query.UpdateCredentialByUserId(tx, ...)
          db.Query.PasswordResetMarkUsed(tx, rst.Pk)

          // Invalidează toate celelalte link-uri de resetare ale user-ului
          db.Query.PasswordResetInvalidateSiblings(tx, ...)

          // Revocă TOATE sesiunile active (Logout de pe toate dispozitivele)
          return db.Query.RevokeUserSessions(tx, ...)
      })
  }
  ```
]
