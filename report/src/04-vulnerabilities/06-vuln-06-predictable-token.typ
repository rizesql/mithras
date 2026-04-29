== VULN-06: Token de Resetare Predictibil

Funcția de recuperare a parolei din `internal/auth/password_reset.go` generează token-ul
pe versiunea vulnerabilă sub forma simplificată `base64(email + unix_timestamp)`.
Suplimentar, endpoint-ul final de consum `/reset-password` nu marchează token-ul ca fiind
utilizat și nu impune nicio restricție de expirare sau invalidare a sesiunilor. Această
defecțiune de logică a fost catalogată ca A07:2025 și CWE-640 (Weak Password Recovery
Mechanism for Forgotten Password) @cwe640.

=== Fundamentare Teoretică și Implementare Defectuoasă

Sistemele de recuperare a parolei acționează ca un canal secundar de autentificare
(out-of-band). Dacă token-ul de resetare livrat prin email poate fi ghicit sau dedus de o
terță parte, securitatea parolei principale devine irelevantă. Utilizarea unui generator
de numere pseudo-aleatoare non-criptografic (PRNG) sau derivarea token-ului din metadate
publice reduce drastic entropia.

Implementarea V1 construiește token-ul direct din timestamp-ul UNIX (rezoluție la nivel de
secundă) concatenat cu adresa de email:

#figure(
  caption: [Implementare V1 - Generare predictibilă a token-ului de recuperare],
)[
  ```go
  // internal/auth/password_reset.go (branch: vulnerable)
  func generateResetToken(email string) string {
      // Bază de entropie critic scăzută: timestamp UNIX în secunde
      now := time.Now().Unix()
      raw := fmt.Sprintf("%s:%d", email, now)

      // Token-ul este o simplă codificare Base64 a unui text previzibil
      return base64.URLEncoding.EncodeToString([]byte(raw))
  }
  ```
]

Suplimentar, la schimbarea efectivă a parolei, sistemul nu marchează token-ul ca fiind
utilizat (`used_at IS NULL` rămâne activ) și nu forțează un logout global (`Force Logout
All`).

=== Mecanism de Atac și Exploatare

Aceasta reprezintă cea mai severă fisură logică a sistemului, permițând o preluare
completă și tăcută a contului (*Account Takeover*). Un atacator poate declanșa cererea de
recuperare a parolei introducând email-ul victimei pe pagina de "Forgot Password".
Înregistrând momentul exact al cererii (într-o fereastră de câteva secunde), atacatorul
poate genera local (offline) toate token-urile posibile codificând Base64
`email + unix_timestamp`. Trimițând imediat o cerere către `POST /reset-password` cu
token-ul calculat, atacatorul modifică parola înainte ca victima să poată deschide
email-ul legitim, preluând instantaneu controlul asupra contului.
