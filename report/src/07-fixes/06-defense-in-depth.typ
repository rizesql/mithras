== Apărare în Adâncime (Defense in Depth)

Arhitectura finală implementată pe branch-ul *main* nu se limitează doar la închiderea
vulnerabilităților punctuale investigate, ci introduce și un set de controale de siguranță
structurale pentru a limita suprafața potențială de expunere la pericole încă necunoscute.

=== Interogări Parametrizate Stricte (sqlc)

În locul utilizării bibliotecilor ORM (*Object-Relational Mapping*) generice sau al
formatărilor manuale de tip string (principala sursă a injecțiilor SQL), interacțiunile cu
baza de date PostgreSQL din cadrul aplicației Mithras sunt generate exclusiv prin
intermediul setului de instrumente `sqlc`.

Această tehnologie pre-procesează fișierele `.sql` strict tipizate, transformând sintaxa
în funcții Go. Astfel, orice variabilă din cod trimisă către PostgreSQL este interpretată
prin protocolul binar (la nivel de socket, folosind marcaje specifice precum `$1`, `$2`),
interzicând complet, din faza de generare a codului, interpretarea malițioasă a corpului
interogării.

=== Prevenirea Vulnerabilităților de Tip IDOR

În endpoint-urile care necesită manipularea obiectelor direct asociate unui profil de
utilizator, sistemul exclude în totalitate utilizarea parametrilor din cererile client
(cum ar fi identificatori trimiși direct sub forma unui atribut JSON în corpul cererii sau
parametri în URL).

Autentificarea și autorizarea operațiunilor sensibile se bazează exclusiv pe identitatea
extrasă și validată server-side. Identificatorul utilizatorului (precum `UserPk` sau
`UserID`) este preluat strict din payload-ul token-ului de acces (JWS), a cărui
integritate este garantată matematic de semnătura asimetrică EdDSA.

Prin refuzul de a acorda încredere parametrilor de identificare furnizați de client (ex.
ID-uri manipulate în corpul cererii sau în URL), sistemul neutralizează structural
întreaga clasă de vulnerabilități bazate pe referințe directe nesigure la obiecte
(IDOR --- *Insecure Direct Object Reference*).
