= Demonstrare Atac (PoC)

Această secțiune detaliază exploatarea practică a fiecărei vulnerabilități documentate
anterior. Toate demonstrațiile sunt executate pe branch-ul `vulnerable` (versiunea V1),
utilizând scripturi ofensive dezvoltate special pentru acest audit (în directorul
`exploits/`) sau utilitare standard de interacțiune HTTP (`curl`).

Conform metodologiei *Build, Hack & Secure*, fiecare atac este prezentat într-un format
uniform, asigurând reproductibilitatea pas cu pas: precondiții, execuție, capturi de ecran
(care includ terminalul și identificarea unică a mediului de lucru) și rezultatul obținut.

#include "01-poc-01-password-policy.typ"
#include "02-poc-02-plaintext-storage.typ"
#include "03-poc-03-rate-limiting.typ"
#include "04-poc-04-user-enumeration.typ"
#include "05-poc-05-insecure-sessions.typ"
#include "06-poc-06-predictable-token.typ"
