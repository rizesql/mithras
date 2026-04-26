= Introducere

Prezentul raport documentează ciclul complet de dezvoltare, auditare și fortificare a
aplicației _Mithras_, un furnizor de identitate (Identity Provider - IdP) dezvoltat ca
studiu de caz pentru disciplina "Dezvoltarea Aplicațiilor Software Securizate". Proiectul
fundamentează tranziția de la o arhitectură inițial vulnerabilă la un sistem rezilient,
operând sub paradigma de securitate _Defense in Depth_ (Apărarea în Adâncime).

Metodologia aplicată urmează etapele unui ciclu de dezvoltare securizat (Secure SDLC),
obiectivul principal fiind reducerea profilului de risc și a suprafeței de atac prin
implementarea unor controale tehnice stratificate. Validarea acestor mecanisme este
realizată prin demonstrarea practică a vectorilor de compromitere (Proof of Concept) pe
versiunea de referință (V1), urmată de o re-testare riguroasă a controalelor de atenuare
integrate în stadiul final al aplicației (V2).

#include "01-description.typ"
#include "02-architecture.typ"
