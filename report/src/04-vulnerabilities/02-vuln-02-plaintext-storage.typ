== VULN-02: Stocare a Parolelor în Clar (Plaintext Storage)

În funcția `Hash()` din `internal/password/hashed.go`, aplicarea algoritmului Argon2id
este suprimată pe versiunea vulnerabilă, funcția returnând valoarea brută a parolei sub
masca tipului `Hashed`. Ca rezultat, coloana `secret` din tabelul `credentials` stochează
parolele în format text simplu. Defectul aparține categoriei A02:2025 (Cryptographic
Failures) și este identificat ca CWE-256 (Plaintext Storage of a Password). @cwe256

=== Fundamentare Teoretică și Implementare Defectuoasă

Principiul *Defense in Depth* dictează că datele de autentificare nu trebuie stocate
niciodată într-o formă reversibilă. Deși accesul la baza de date ar trebui să fie strict
controlat, un sistem matur presupune compromiterea eventuală a stratului de persistență.
Fără o funcție de derivare a cheii (*Key Derivation Function* - KDF) adaptată
hardware-ului, precum Argon2id, secretele sunt compromise automat odată cu baza de date.

Pe branch-ul vulnerabil, logica de hashing criptografic a fost înlocuită cu o funcție
passthrough:

#figure(
  caption: [Implementare V1 - Suprimarea hashing-ului Argon2id],
)[
  ```go
  // internal/password/hashed.go (branch: vulnerable)
  func (r Raw) Hash() (Hashed, error) {
      // Hashing-ul Argon2id este dezactivat intenționat.
      // Parola în clar este returnată ca atare și va fi salvată în
      // coloana `secret` din baza de date.
      return Hashed{ value: r.value }, nil
  }
  ```
]

=== Mecanism de Atac și Exploatare

Dacă un atacator obține acces de citire la nivelul bazei de date (printr-un backup
neprotejat, o eroare de configurare în mediul cloud sau o injecție SQL laterală), efortul
de preluare a conturilor este nul. Parolele sunt extrase direct din baza de date, fără
necesitatea alocării de resurse computaționale (GPU/ASIC) pentru spargere. Această breșă
permite compromiterea încrucișată a victimelor pe alte platforme (*Credential Reuse*),
extinzând impactul atacului mult dincolo de granițele sistemului Mithras.
