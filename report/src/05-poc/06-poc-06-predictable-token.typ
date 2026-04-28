== PoC --- VULN-06: Previzionarea Token-ului de Resetare Parolă

Această procedură dovedește impactul catastrofal al token-urilor generate
lipsite de entropie, ducând la un *Account Takeover* total.

*Precondiție:* Cunoaștem o adresă de email legitimă a unui utilizator pe care dorim
să-l vizăm (ex. `victim@authx.com`). Baza de date conține înregistrarea pentru acesta.

*Pași de execuție:*
Utilizăm script-ul avansat `exploits/password-reset/main.go`, creat special
pentru a automatiza ciclul complet de cerere-previzionare-resetare a parolei.

```sh
cd exploits/password-reset
go run . -email victim@authx.com -password "HackerOwned1!"
```

Scriptul parcurge trei etape:
1. Emite cererea `POST /forgot-password` inițiind procesul de recuperare din partea
  serverului.
2. În funcție de timestamp-ul curent, calculează algoritmul slab de codare
  `base64(email + unix_timestamp)` pentru a obține secretul. Citim partea ID din
  log-urile server-ului (simulând o vulnerabilitate secundară de obținere a ID-ului
  sau un atac de timp restrâns) și formatăm tokenul.
3. Apelează instantaneu `POST /reset-password` trimițând predicția împreună cu noua
  parolă pe care o setăm pentru victimă.

*Rezultat observat:*
Atacatorul nu are nevoie să acceseze căsuța de email a victimei. Scriptul confirmă
"Prediction matches the server-generated secret!". Ulterior, trimite predicția pe post
de token de resetare, iar serverul răspunde cu 200 OK. Parola este setată la
`HackerOwned1!`, și preluarea de cont a reușit.

#figure(
  caption: [Demonstrația vulnerabilității VULN-06: Previzionarea cu succes a token-ului de resetare bazat pe adresa de email și timestamp.],
  image("assets/06-poc-06.png"),
)
