== VULN-01: Politică de Parole Absentă (Weak Password Requirements)

Constructorul entității `password.New()` din fișierul `internal/password/raw.go`
(versiunea V1) acceptă orice șir de caractere non-gol, validările pentru lungimea minimă
și complexitate fiind complet eliminate. Defectul este clasificat conform OWASP Top 10 ca
A07:2025 (Identification and Authentication Failures) și CWE-521 (Weak Password
Requirements) @cwe521.

=== Fundamentare Teoretică și Implementare Defectuoasă

O politică de parole robustă reprezintă primul strat de apărare împotriva preluării
conturilor. În absența impunerii unei entropii minime (lungime și varietate a setului de
caractere), utilizatorii au tendința de a alege parole scurte, previzibile sau derivate
din informații personale.

În versiunea vulnerabilă a sistemului Mithras, constrângerile au fost eliminate deliberat
din faza de instanțiere a valorii brute:

#figure(
  caption: [Implementare V1 - Suprimarea validării parolei],
)[
  ```go
  // internal/password/raw.go (branch: vulnerable)
  func New(value string) (Raw, error) {
      if strings.TrimSpace(value) == "" {
          return Raw{}, errors.New("password cannot be empty")
      }

      // Validările de lungime (min 8) și complexitate (majusculă, cifră, special)
      // sunt dezactivate intenționat pe acest branch.
      // Sistemul va accepta parole triviale (ex. "a", "123").

      return Raw{value: value}, nil
  }
  ```
]

=== Mecanism de Atac și Exploatare

Existența acestei vulnerabilități anulează rezistența sistemului la atacuri de tip
*Credential Stuffing* sau *Brute-Force*. Chiar și în prezența unor mecanisme moderne de
hashing, o parolă lipsită de entropie poate fi spartă instantaneu. Un atacator poate
utiliza dicționare comune (precum `rockyou.txt`) pentru a efectua încercări de
autentificare. Sistemul va accepta la înregistrare parole de un singur caracter, creând o
vulnerabilitate exploatabilă ulterior la nivelul endpoint-ului de login.
