== Remediere VULN-01 & VULN-02: Politică de Parole și Hashing Argon2id

Defectele fundamentale privind acceptarea parolelor triviale și stocarea lor în clar
au fost remediate prin impunerea unei politici stricte și utilizarea algoritmului Argon2id.

=== Politica de Complexitate

Constructorul parolei din `internal/password/raw.go` a fost refactorizat pentru a impune
o lungime minimă de 8 caractere și prezența simultană a literelor mari, mici, a cifrelor
și a caracterelor speciale. Orice intrare care nu respectă aceste criterii este respinsă,
prevenind la sursă vulnerabilitatea de tip CWE-521.

=== Derivarea Cheii cu Argon2id

Pentru remedierea CWE-256 (stocarea în clar), funcția de transformare utilizează Argon2id
cu parametri adaptați mediului de producție ($m=65536, t=2, p=2$). Argon2id a fost ales în
detrimentul algoritmilor rapizi (ex. SHA-256) datorită factorului său de întărire pe baza
consumului de memorie (*Memory Hardness*), care limitează sever eficiența atacurilor
paralele executate pe GPU sau ASIC.

Valoarea stocată în baza de date este codificată în formatul standard PHC, care
încapsulează inclusiv salt-ul de 16 octeți generat aleatoriu. Astfel, parolele stocate
ireversibil nu pot fi compromise în cazul scurgerilor stratului de persistență.
