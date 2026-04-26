== VULN-03: Lipsă de Rate Limiting pe Autentificare

Pe branch-ul vulnerabil, middleware-ul de rate limiting a fost complet eliminat din fluxul
cererilor HTTP aferente endpoint-ului de autentificare. Orice interfață de contorizare
sau frânare temporală este absentă, permițând procesarea neîngrădită a cererilor. Aceasta
reprezintă o vulnerabilitate critică, mapată ca A07:2025 și CWE-307 (Improper Restriction
of Excessive Authentication Attempts) @cwe307.

=== Fundamentare Teoretică și Implementare Defectuoasă

Securitatea sistemelor de autentificare moderne se bazează adesea pe factorul timp.
Limitarea numărului de cereri per IP sau per cont (Rate Limiting / Account Lockout)
transformă un atac de forță brută teoretic posibil într-o operațiune ineficientă practic.
Fără această frânare, un atacator poate trimite mii de cereri pe secundă.

În versiunea V1, logica de rutare elimină cu desăvârșire orice validare de frecvență a
cererilor:

#figure(
  caption: [Implementare V1 - Eliminarea middleware-ului de Rate Limiting],
)[
  ```go
  // internal/mithras/routes/routes.go (branch: vulnerable)
  // ...
  srv.RegisterRoute(login.New(plt),
      withPanicRecovery,
      withTimeout,
      // withRateLimit,
      // login.RateLimit(plt), -> Omiterea deliberată a limitării stricte per cont
      withValidation,
  )
  ```
]

=== Mecanism de Atac și Exploatare

Endpoint-ul `/login` devine o țintă directă pentru scripturi de automatizare. Fără nicio
penalizare temporală sau blocare la nivel de cont, un atacator poate rula un atac de
dicționar exhaustiv. Aceasta transformă forța brută dintr-o procedură costisitoare într-una
rapidă și asimptotic garantată de succes, limitată exclusiv de latența conexiunii la rețea
și de resursele alocate serverului IdP. Această vulnerabilitate este exacerbată sever în
combinație cu politica slabă de parole (VULN-01).
