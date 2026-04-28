== VULN-04: Enumerare Utilizatori (User Enumeration)

Mecanismul de autentificare al utilizatorilor din `internal/auth/login.go` răspunde
diferențiat pe versiunea V1: un utilizator inexistent generează eroarea `errUserNotFound`,
pe când o parolă greșită pentru un utilizator valid generează `errWrongPassword`.
Suplimentar, funcția omite calculul pe un *dummy hash* pentru utilizatorii inexistenți,
generând o discrepanță majoră de timp de răspuns. Vulnerabilitatea se încadrează la
A07:2025 și CWE-204 (Observable Response Discrepancy) @cwe204.

=== Fundamentare Teoretică și Implementare Defectuoasă

Atacurile prin canale laterale (*Side-Channel Attacks*) exploatează modul de funcționare a
hardware-ului sau a software-ului, nu defectele structurale ale protocolului. Un *Timing
Attack* se bazează pe faptul că operațiunile criptografice (precum Argon2id) consumă timp
măsurabil. Dacă sistemul sare peste calculul Argon2id atunci când adresa de email nu
există în baza de date, sistemul returnează răspunsul instantaneu.

În versiunea V1, răspunsul este diferențiat atât la nivel semantic, cât și temporal:

#figure(
  caption: [Implementare V1 - Diferențierea mesajelor și evitarea calculului pe dummy hash],
)[
  ```go
  // internal/auth/login.go (branch: vulnerable)
  func (a *Auth) Login(email, password string) error {
      usr, err := a.db.GetUserByEmail(ctx, email)
      if err != nil {
          // Diferență semantică: returnăm instantaneu
          return errUserNotFound
      }

      match, err := verifyPassword(password, user.Secret)
      if !match {
          return errWrongPassword // Diferență semantică
      }
      return nil
  }
  ```
]

=== Mecanism de Atac și Exploatare

Prin analiza sintactică a corpului de răspuns JSON (eroare explicită) sau prin profilarea
statistică a temporizării cererii, un atacator remote neautentificat poate sonda masiv
endpoint-ul `/login`. El poate rula un dicționar de adrese de email populare, construind
rapid o listă precisă de utilizatori existenți pe platforma Mithras. Această listă este
apoi folosită pentru a focaliza atacurile de forță brută (VULN-03) sau pentru campanii de
spear-phishing direcționate, reducând "zgomotul" și sporind eficiența compromiterii.
