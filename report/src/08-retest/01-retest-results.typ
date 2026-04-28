
== VULN-01: Înregistrare cu parolă trivială

La re-executarea script-ului `exploits/weak-password` cu payload-ul inițial (parola `"1"`),
sistemul respinge imediat cererea la nivelul validării datelor de intrare.

*Rezultat observat:* Răspuns HTTP `400 Bad Request` cu un mesaj de eroare comprehensiv.
Politicile de complexitate implementate în constructorul `password.New()` interceptează
încercarea de instanțiere a unei structuri `Raw` necorespunzătoare. Cererea HTTP este
terminată cu eroare înainte ca fluxul de execuție să atingă logica de creare a utilizatorului
în tranzacția PostgreSQL.

#figure(
  caption: [Retest VULN-01: Cererea cu parolă trivială este respinsă de server cu eroarea HTTP 400 Bad Request.],
  image("assets/01-poc-01.png"),
)

== VULN-02: Extragere Bazei de Date

Interogarea SQL directă pe tabelul `credentials` nu mai expune nicio parolă în clar,
diminuând drastic valoarea datelor exfiltrate.

*Rezultat observat:* Coloana `secret` conține exclusiv hash-uri securizate în format PHC,
precedate de `$argon2id$`, integrând salt-uri unice (16 octeți) pentru fiecare intrare.
Spre deosebire de un text simplu care poate fi reutilizat imediat pe alte platforme,
reconstituirea unei singure parole din acest dump ar necesita un efort computațional masiv
(ASIC/GPU), dictat de memoria impusă de parametrii Argon2id ($m=65536$). Atacul devine
astfel nefezabil din punct de vedere economic și temporal.

#figure(
  caption: [Retest VULN-02: Parola este stocată securizat în format PHC sub formă de hash Argon2id cu salt.],
  image("assets/02-poc-02.png"),
)

== VULN-03: Brute Force

La rularea script-ului `exploits/brute-force`, mecanismul de rate limiting intervine prompt,
frânând și blocând execuția automată a cererilor.

*Rezultat observat:* După un număr finit de încercări secvențiale eșuate, script-ul
primește răspunsul HTTP `429 Too Many Requests`.

#figure(
  caption: [Retest VULN-03: Blocarea atacului forță brută de mecanismul de rate limiting, care returnează HTTP 429 Too Many Requests.],
  image("assets/03-poc-03.png"),
)

== VULN-04: Enumerare Utilizatori

Sondarea conturilor, existente sau inexistente, prin `exploits/user-enumeration` generează
acum comportamente identice pe versiunea securizată.

*Rezultat observat:* Ambele cereri returnează uniform eroarea semantică
`{"detail":"Invalid email or password."}`, eliminând posibilitatea filtrării pe baza
textului primit. Suplimentar, analiza timpului de răspuns demonstrează eficacitatea
mecanismului de *dummy hash*. Deoarece serverul execută funcția Argon2id `verifyPassword`
chiar și atunci când interogarea `GetUserByEmail` eșuează (utilizând hash-ul fictiv),
ambele cereri consumă aproximativ același timp de procesare CPU. Această uniformizare
temporală elimină complet oracolul bazat pe latență, blocând faza de recunoaștere a atacului.

#figure(
  caption: [Retest VULN-04: Eliminarea oracolului de enumerare prin returnarea aceluiași mesaj de eroare și a unui timp de execuție uniform.],
  image("assets/04-poc-04.png"),
)

== VULN-05: Furtul Sesiunii (XSS)

Simularea unei vulnerabilități Cross-Site Scripting (XSS) în browser nu mai compromite
integritatea autentificării curente.

*Rezultat observat:* Executarea `document.cookie` în consola browserului pentru
versiunea securizată returnează un string gol (sau doar cookie-uri tehnice, non-sensibile).
Token-ul de acces este invizibil pentru motorul JavaScript, protecția fiind garantată
la nivelul clientului web de flag-ul `HttpOnly`.

Chiar dacă atacatorul ar reuși să intercepteze token-ul JWS printr-un atac avansat
la nivel de rețea (ex. lipsa flag-ului `Secure` peste o conexiune HTTP), durata de
viață redusă la 5 minute și revocarea instantanee a token-ului de refresh (*Single-Use*)
limitează fereastra de oportunitate la o perioadă infimă.

#figure(
  caption: [Retest VULN-05: Prevenirea furtului de sesiune prin ascunderea token-ului de acces cu ajutorul flag-ului HttpOnly.],
  image("assets/05-poc-05.png"),
)

== VULN-06: Token de Resetare Predictibil

Reluarea atacului prin script-ul `exploits/password-reset` pe versiunea securizată
demonstrează că algoritmul de generare a token-ului nu mai poate fi spart prin ghicirea secretului.

*Rezultat observat:* Scriptul calculează o predicție bazată pe algoritmul vechi (timestamp și email).
Totuși, la introducerea token-ului real generat de sistemul modern (preluat din log-uri),
scriptul raportează imediat eșecul atacului: `Prediction did not match!`.

Sistemul modern a schimbat complet formatul într-o structură compusă `{id}.{secret}`,
iar secretul este generat folosind `crypto/rand` (256 biți de entropie). Oricât
de multe încercări ar face un atacator offline, predicția sa (ex: `cml6ZXNxb...`) va diferi
complet de valoarea criptografică reală (ex: `_bt_YH4...`).

Deși scriptul trimite mai departe token-ul valid introdus manual și primește un răspuns
`200 OK` (simulând astfel un flux legitim de resetare a parolei de către un utilizator real),
scopul atacului de *a ghici* token-ul fără acces la email-ul victimei a eșuat irevocabil.

#figure(
  caption: [Retest VULN-06: Demonstrarea eșecului de predicție a token-ului ("Prediction did not match!"), urmată de o resetare legitimă folosind token-ul real din log-uri.],
  image("assets/06-poc-06.png"),
)
