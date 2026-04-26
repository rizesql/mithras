== Înregistrare Utilizator (Register)

Procesul de înregistrare constituie prima linie de apărare a sistemului, fiind
responsabil pentru validarea identității utilizatorului și transformarea ireversibilă a
credențialelor înainte de stocare. Endpoint-ul `POST /register` implementează un flux
riguros de validare și persistență, asigurând integritatea datelor și reziliența
împotriva atacurilor de tip brute-force offline.

=== Validarea și Sanitizarea

Mithras utilizează specificația *OpenAPI 3.1* și `middleware.WithValidation` pentru a
asigura validarea strictă a cererilor HTTP la nivel de endpoint. Acest mecanism garantează că
orice corp JSON este bine format, nu conține câmpuri necunoscute și respectă constrângerile
de tip (ex. adrese de email valide). Orice cerere incorectă este respinsă imediat,
înainte de a atinge logica de business.

Această abordare permite ca implementarea internă din `internal/auth/register.go` să fie
"curată" și concentrată pe securitate, operând sub asumpția că datele de intrare sunt
deja validate structural. Acest lucru elimină clase întregi de vulnerabilități, precum
cele legate de procesarea necorespunzătoare a input-ului sau erorile de tip.
Sistemul impune suplimentar o politică de parole robustă, implementată în constructorul
`password.New()` din `internal/password/raw.go`. Această abordare garantează că un
obiect de tip `password.Raw` (parolă validă structural) nu poate fi instanțiat dacă nu
îndeplinește criteriile minime de securitate, precum lungimea minimă de 8 caractere
și cerințele de complexitate (cel puțin o literă mare, o literă mică, o cifră și
un caracter special). Validarea email-ului este realizată prin pachetul
`internal/email`, care verifică formatul RFC 5322 @rfc5322, prevenind injecțiile
de caractere invalide în baza de date.

=== Hashing Argon2id

Mithras utilizează algoritmul *Argon2id* @rfc9106 pentru protecția parolelor, fiind
standardul actual recomandat de OWASP @owasp2025. Parametrii utilizați în versiunea
securizată includ o memorie de 64 MiB ($m=65536$), 2 iterații temporale ($t=2$) și
un factor de paralelism de 2 ($p=2$). Salt-ul este format din 16 octeți generați
aleator, rezultând o cheie derivată de 32 octeți.

Alegerea Argon2id în detrimentul bcrypt sau PBKDF2 este motivată de reziliența sa
împotriva atacurilor accelerate prin hardware (GPU/ASIC), datorită parametrului de
cost al memoriei care forțează atacatorul să aloce resurse hardware semnificative per
încercare.

#figure(
  caption: [Implementare Argon2id în `internal/password/hashed.go`],
)[
  ```go
  func (r Raw) Hash() (Hashed, error) {
      salt := make([]byte, saltLen)
      if _, err := rand.Read(salt); err != nil {
      	return Hashed{}, err
      }

      hash := argon2.IDKey(
          []byte(r.value), salt,
          argon2id.time, argon2id.memory, argon2id.threads, keyLen,
      )

      // Codificat in format PHC
      return Hashed{value: encoded}, nil
  }
  ```
]

=== Fluxul de Înregistrare

Arhitectura procesului de înregistrare pune accent pe atomicitatea operațiunilor și
pe izolarea materialului criptografic. Fluxul de date între Consumer (BFF), Mithras
IdP și baza de date este ilustrat în figura de mai jos:

#figure(
  caption: [Diagramă de Secvență: Fluxul de Înregistrare Utilizator],
  include "assets/01-register-diagram.typ",
) <fig-register-flow>

=== Persistență Atomică și Gestiunea Identității

Stocarea datelor de înregistrare este realizată într-o singură tranzacție PostgreSQL
pentru a preveni stările inconsistente. Pe lângă crearea profilului de bază, Mithras
implementează două mecanisme critice pentru un sistem de autentificare comprehensiv. Primul
mecanism vizează *Istoricul Parolelor*, unde hash-ul inițial este inserat și în
tabelul `password_history`. Acest lucru permite impunerea unor politici viitoare
de rotație, împiedicând utilizatorii să revină la parole utilizate anterior, chiar
dacă acestea au fost sigure la momentul respectiv. Al doilea mecanism este
*Atribuirea Rolurilor (RBAC)*, prin care fiecare utilizator este înregistrat
implicit cu rolul `USER` în tabelul `roles`. Această legătură atomică asigură că nu
există identități "orfane" în sistem care ar putea obține privilegii neprevăzute
prin erori de logică în codul de autorizare.

În cazul în care adresa de email este deja înregistrată, sistemul detectează eroarea de
constrângere `UNIQUE` și returnează o eroare explicită (`errDuplicateEmail`). Deși
această decizie permite teoretic enumerarea utilizatorilor la înregistrare, ea
reprezintă un compromis asumat (*Security vs. Usability Trade-off*). Opacizarea
completă a acestui endpoint ar genera un volum ridicat de incidente de suport
(utilizatori care nu înțeleg de ce nu primesc email-ul de confirmare). Pentru a
mitiga acest risc, s-a optat pentru menținerea opacității stricte la nivelul
endpoint-urilor de autentificare (`/login`) și recuperare (`/forgot-password`),
în timp ce endpoint-ul `/register` este protejat defensiv prin aplicarea unui
mecanism de *Rate Limiting* agresiv bazat pe adresa IP, făcând atacurile de
enumerare la scară largă statistic inaplicabile.
