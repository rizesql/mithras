= Prezentare Vulnerabilități: OWASP Mapping

Această secțiune detaliază profilul de vulnerabilitate al sistemului, analizând defecțiunile de
proiectare și implementare introduse deliberat pe branch-ul `vulnerable` (versiunea V1).
Obiectivul este de a demonstra cum o eroare structurală la nivel de cod conduce la o
breșă exploatabilă în contextul mecanismelor de autentificare.

Fiecare vulnerabilitate este mapată conform standardului *OWASP Top 10* @owasp2025 și
sistemului de clasificare *CWE* (Common Weakness Enumeration) @cwe_framework, oferind o fundamentare
obiectivă a riscului. Aceste vulnerabilități nu mai sunt prezente în starea curentă a
proiectului (versiunea V2), fiind neutralizate prin hardening-ul detaliat în secțiunea de
implementare.

#include "01-vuln-01-password-policy.typ"
#include "02-vuln-02-plaintext-storage.typ"
#include "03-vuln-03-rate-limiting.typ"
#include "04-vuln-04-user-enumeration.typ"
#include "05-vuln-05-insecure-sessions.typ"
#include "06-vuln-06-predictable-token.typ"
