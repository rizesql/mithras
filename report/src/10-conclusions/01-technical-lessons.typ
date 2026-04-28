== Lecții Tehnice

Analiza vectorilor de atac asupra mecanismelor de autentificare a demonstrat
viabilitatea următoarelor concepte tehnice.

=== Rata de Răspuns a Sistemului ca Interfață de Expunere
Vulnerabilitatea de enumerare a utilizatorilor (VULN-04) demonstrează că un mesaj de eroare
generic este o măsură necesară, dar insuficientă, în lipsa unei uniformizări temporale.
Diferența de timp de răspuns între evaluarea unui utilizator inexistent și rularea
algoritmului Argon2id pentru un utilizator valid a oferit atacatorului un canal
lateral (*Side-Channel*) la fel de precis ca o confirmare explicită. Introducerea
mecanismului de *dummy hash* demonstrează că, în contextul securității ofensive, timpul de
procesare reprezintă o scurgere de informații (*Information Leak*).

=== Complexitatea Algoritmică vs. Securitate Efectivă
Metoda de generare a token-ului de resetare de pe branch-ul vulnerabil implică formatări,
concatenări de variabile și codificări Base64 (`base64(email + unix_timestamp)`), rezultând
într-o structură aparent complexă. Cu toate acestea, oferă zero protecție criptografică. În
contrast, apelarea directă a funcției `crypto/rand` într-un buffer de 32 de octeți este
trivială din punct de vedere al implementării, dar computațional infezabil de dedus. Lecția
extrasă este că securitatea nu decurge din complexitatea transformărilor textuale, ci
exclusiv din entropia sursei generatoare.

=== Defense in Depth pentru Vectori Nedescoperiți
Utilizarea cookie-urilor de sesiune fără flag-ul `HttpOnly` (VULN-05) nu generează un atac
direct în absența unei vulnerabilități de tip Cross-Site Scripting (XSS). Totuși, paradigma
apărării în adâncime (*Defense in Depth*) impune implementarea unor controale de mitigare
chiar și pentru vectorii absenți momentan. Setarea corectă a cookie-ului asigură că, în
eventualitatea apariției unei breșe în stratul de frontend, materialul criptografic
esențial rămâne izolat față de motorul de execuție JavaScript.

=== Validarea Exclusiv Server-Side a Identității
Protejarea sistemului împotriva referințelor directe la obiecte nesigure (IDOR) a impus o
regulă strictă: datele transmise de client nu au autoritate de decizie. Indiferent de ID-urile
oferite într-un payload JSON sau ca parametri URL, aplicația trebuie să se bazeze exclusiv pe
identitatea extrasă din token-ul de acces semnat asimetric (JWS EdDSA).
