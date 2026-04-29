== Remediere VULN-05: JWS Stateless și Securizarea Cookie-urilor

Pentru a neutraliza vulnerabilitățile de expunere la nivelul sesiunii (CWE-614) a fost
aplicată o strategie combinată între stratul de consum (BFF) și serverul IdP.

=== Cookie-uri Securizate

Sistemul BFF setează exclusiv cookie-ul de sesiune cu întregul set de flag-uri
obligatorii: `HttpOnly` (care interzice total interacțiunea cu API-ul `document.cookie`
din mediul JavaScript, blocând direct exfiltrarea furtului de sesiune prin atacuri XSS),
`Secure` (garantând transmiterea cookie-ului doar prin conexiuni criptate TLS), și
`SameSite=Strict` (o protecție by-design foarte puternică împotriva vulnerabilităților
CSRF).

=== Rotația Token-ului și Fereastra Minimă de Expunere

La nivel de backend, durabilitatea sesiunii a fost reproiectată. Durata de viață a
token-ului de acces asimetric (JWS) a fost redusă de la 24 de ore la doar 5 minute. Pentru
a nu degrada experiența utilizatorului prin cerințe repetate de logare, un *Refresh Token*
opac (stocat criptat ca `SHA-256` în PostgreSQL) operează în fundal, respectând o politică
strictă de *rotație la fiecare utilizare* (Single-Use).

Suplimentar, operațiunea de logout (`internal/auth/logout.go`) include comanda esențială
`RevokeUserSessions`, setând marcajul `revoked_at` pentru sesiunile din baza de date.
Astfel, reducând fereastra activă a JWS-ului la 5 minute, intervalul în care o sesiune
potențial capturată înainte de deconectare poate fi valorificată devine minim.
