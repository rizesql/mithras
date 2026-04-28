#import "lib/layout.typ": ieee
#import "@preview/codly:1.3.0": *
#import "@preview/codly-languages:0.1.1": *

#set text(lang: "ro")

#show: ieee.with(
  title: [Mithras Identity Provider: Raport de Audit și Securizare],
  authors: (
    (
      name: "Rizescu Iulian Ștefan",
      email: "iulian-stefan.rizescu@s.unibuc.ro",
    ),
  ),

  // abstract: [
  //   Acest raport detaliază procesul de proiectare, implementare și securizare a sistemului
  //   *Mithras*, un furnizor de identitate (Identity Provider - IdP) bazat pe protocolul
  //   OAuth 2.0. Lucrarea urmărește metodologia Secure SDLC, prezentând tranziția tehnică
  //   de la o implementare inițial vulnerabilă (V1) la o arhitectură fortificată (V2) prin
  //   aplicarea controalelor de securitate recomandate de @owasp2025. Sunt identificate,
  //   demonstrate practic și remediate șase clase de vulnerabilități: politică de parole
  //   absentă, stocare a parolelor în clar, lipsă de rate limiting, enumerare utilizatori,
  //   gestionare nesigură a sesiunilor și token de resetare predictibil. Remedierile
  //   implementate includ hashing Argon2id @rfc9106, semnare token JWS cu EdDSA @rfc8032
  //   și rate limiting prin token bucket Redis, validate prin re-testare practică.
  // ],
  bibliography: bibliography("refs.bib"),

  cols: 2,
  figure-supplement: [Fig.],
)

#show: codly-init.with()
#codly(
  languages: codly-languages,
  zebra-fill: none,
  display-icon: false,
)

#include "01-introduction/00-index.typ"
#include "02-setup/00-index.typ"
#include "03-implementation/00-index.typ"
#include "04-vulnerabilities/00-index.typ"
#include "05-poc/00-index.typ"
#include "06-impact/00-index.typ"
#include "07-fixes/00-index.typ"
#include "08-retest/00-index.typ"
#include "09-audit/00-index.typ"
#include "10-conclusions/00-index.typ"
#include "11-bonus/00-index.typ"
