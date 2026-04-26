== Remediere VULN-04: Uniformizarea Răspunsurilor și Dummy Hash

Vulnerabilitatea de enumerare a utilizatorilor (CWE-204) a necesitat o mitigare
bidimensională, vizând atât uniformizarea semanticii răspunsurilor HTTP, cât și
uniformizarea amprentei temporale a cererii.

=== Răspuns Semantic Uniform

Funcția de autentificare din `internal/auth/login.go` a fost ajustată pentru a masca
intern motivul eșecului. Se returnează o singură eroare generică (`errInvalidCredentials`,
transformată la exterior în HTTP 401 Unauthorized) absolut identică și pentru situația în
care adresa de email nu există în sistem, și pentru cazurile în care parola este incorectă.

=== Atenuarea Atacurilor de Profilare Temporală (*Timing Attacks*)

Pentru a rezolva discrepanța majoră de execuție descoperită la nivelul analizei (VULN-04),
a fost introdus mecanismul defensiv de *Dummy Hash*. Dacă un utilizator nu este găsit în
baza de date, sistemul nu returnează eroarea instantaneu. În schimb, apelul funcției de
verificare Argon2id se execută integral pe un hash fictiv, pre-generat și menținut în
memorie în timpul funcționării serverului.

În consecință, timpul de răspuns la un eșec (fie email greșit, fie parolă greșită) se menține
relativ constant. Din perspectiva atacatorului care vizează o conexiune de rețea peste
internet, variația introdusă este anulată de instabilitatea rețelei, anulând valoarea
informațională a atacului de timing.

În plus, comparația finală între hash-ul calculat și cel extras din baza de date se face
folosind strict funcția `subtle.ConstantTimeCompare()`, care previne deducția bit-cu-bit.
