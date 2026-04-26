= Implementare MVP. Mecanisme de Autentificare <3-implementation>

Această secțiune documentează arhitectura și implementarea mecanismelor de securitate
din versiunea de referință a sistemului (*v2 - main*). Această versiune constituie
obiectivul final al procesului de securizare, integrând controalele tehnice necesare
pentru a neutraliza vectorii de atac descriși ulterior în acest raport.

Implementarea urmează principiile *Secure by Design* și *Defense in Depth*, tratând
fiecare etapă a fluxului de autentificare ca pe un punct critic de control. Sunt
prezentate detaliile tehnice pentru înregistrarea utilizatorilor, procesul de login cu
protecție împotriva atacurilor de tip brute-force și timing, precum și gestionarea
securizată a sesiunilor prin token-uri cu rotație.

#include "01-register.typ"
#include "02-login.typ"
#include "03-logout.typ"
#include "04-password-reset.typ"
#include "05-sessions.typ"
#include "06-oauth2.typ"
