== Remediere VULN-06: Token de Resetare Criptografic

Prevenirea preluării abuzive a conturilor prin inferența mecanismelor de recuperare
(CWE-640) a presupus înlocuirea completă a token-urilor predictibile cu resurse opace,
înalt-entropice.

=== Generare Sigură Criptografic

Pe versiunea securizată a sistemului, procesul de generare (`base64(email + unix_time)`)
a fost șters. Tokenul de refresh este generat ca o secvență aleatoare de 32 de octeți,
utilizând un generator de numere pseudo-aleatoare securizat criptografic (`crypto/rand`).
Această valoare asigură o entropie de 256 de biți, făcând atacurile de tip brute-force sau
deducerea secretului computațional imposibile.

=== Hashing Unidirecțional și Single-Use

Apărarea în adâncime a fost completată prin modul de stocare a acestui material sensibil.
IdP-ul nu salvează valoarea propriu-zisă a token-ului de resetare, ci amprenta *SHA-256*
a acestuia. Această precauție implică faptul că, și în cazul compromiterii totale a bazei
de date (similar efectului extrem al VULN-02), atacatorul nu poate prelua valorile și
nici genera adrese de resetare funcționale pentru utilizatori.

Pentru a se respecta cerințele de siguranță aplicate asupra acestui flux, utilizarea unui
token validează o secvență atomică de comenzi la nivelul PostgreSQL: marchează secretul
curent drept epuizat (`used_at = now()`), elimină proactiv orice alte apeluri redundante
ale aceluiași utilizator și declanșează *revocarea globală a sesiunilor active* anterior.
Astfel, la setarea noii parole sigure, controlul exclusiv este restaurat imediat pentru
proprietarul real al contului.
