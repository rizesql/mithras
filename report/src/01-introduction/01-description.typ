== Descrierea Aplicației

=== Scopul Pedagogic și Obiectivele Auditului

Demersul practic vizează analiza modului în care mecanismele de autentificare pot fi
compromise în absența unor controale riguroase și fundamentarea tehnică a soluțiilor de
remediere. Procesul necesită adoptarea unui rol dual. În calitate de _Security Auditor_, se
urmărește identificarea vectorilor de atac și demonstrarea impactului acestora asupra
confidențialității, integrității și disponibilității datelor. Complementar, rolul de
_Security Engineer_ presupune implementarea controalelor de securitate compensatorii la
nivelul codului sursă pentru neutralizarea vulnerabilităților identificate.

Sfera auditului se limitează strict la componentele critice ale infrastructurii de tip IAM
(Identity and Access Management): logica de autentificare, persistența și derivarea
cheilor criptografice (Key Derivation), gestiunea stării sesiunilor și protocoalele de
recuperare a accesului.

=== Model de Amenințare și Perimetru

Actorul de amenințare vizat este un atacator extern, neautentificat, cu acces la nivel
de rețea la endpoint-urile HTTP publice ale sistemului. Nu se presupune acces fizic la
infrastructură, privilegii de sistem sau cunoașterea prealabilă a credențialelor.

Perimetrul auditului este definit explicit. Componentele *in-scope* cuprind înregistrarea
utilizatorilor (`/register`), autentificarea (`/login`), deconectarea (`/logout`), rotația
token-urilor (`/token`), resetarea parolei (`/forgot-password`, `/reset-password`) și
gestionarea sesiunilor active. Sunt considerate *out-of-scope* logica de autorizare la nivelul
resurselor de business, vulnerabilitățile IDOR în afara fluxurilor de autentificare,
precum și injecțiile SQL în endpoint-uri nelegate de autentificare.

Tabelul de mai jos listează cele șase clase de vulnerabilități demonstrate, mapate pe
categoriile OWASP Top 10 @owasp2025 și identificatorii CWE corespunzători.

#figure(
  caption: [Vulnerabilități demonstrate],
  table(
    columns: (auto, 1fr, auto, auto),
    table.header[*ID*][*Vulnerabilitate*][*OWASP A:2025*][*CWE*],
    [VULN-01], [Password Policy slab], [A07], [CWE-521],
    [VULN-02], [Stocare nesigură a parolelor], [A02], [CWE-256],
    [VULN-03], [Brute force / lipsă rate limiting], [A07], [CWE-307],
    [VULN-04], [User Enumeration], [A06], [CWE-204],
    [VULN-05], [Gestionare nesigură a sesiunilor], [A07], [CWE-614],
    [VULN-06], [Token de resetare predictibil], [A07], [CWE-640],
  ),
)

=== Scenariul Business: Cazul "AuthX"

Din punct de vedere al contextului de utilizare, sistemul deservește entitatea fictivă
"AuthX", care folosește aplicația pentru medierea accesului angajaților la resurse
corporative sensibile. Premisa auditului este că sistemul se află deja în producție, însă
o evaluare internă a semnalat deficiențe critice la nivelul autentificării. Trecerea la
standarde moderne, precum _Argon2id_ @rfc9106 și _EdDSA_ @rfc8032, răspunde cerinței
imperative de a proteja materialul de autentificare împotriva atacurilor de tip
_offline brute-force_ și a furturilor de date.
