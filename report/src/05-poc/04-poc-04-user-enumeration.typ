== PoC --- VULN-04: Enumerare Utilizatori (User Enumeration)

Demonstrația arată cum mesajele de eroare explicite și variațiile temporale dezvăluie
starea conturilor.

*Precondiție:* Se alege o adresă existentă (`victim@authx.com`) și una inexistentă
(`nobody@authx.com`). Sistemul este pe branch-ul `vulnerable`.

*Pași de execuție:*
Utilizăm script-ul `exploits/user-enumeration/main.go` pentru a efectua două login-uri
paralele și a analiza răspunsurile HTTP semantice returnate de server.

```sh
cd exploits/user-enumeration
go run . -email victim@authx.com
go run . -email nobody@authx.com
```

*Rezultat observat:*
Pentru `victim@authx.com` (existent), serverul returnează în mod clar corpul de eroare:
`{"error": "Incorrect password"}`. Răspunsul este relativ lent datorită evaluării de
memorie necesare pentru procesarea parolei (dacă VULN-02 nu era activ, sau datorită
preluării profilului din baza de date).

Pentru `nobody@authx.com` (inexistent), serverul întoarce: `{"error": "User not found"}`,
iar comparația de parolă se omite, făcând durata de procesare observabil mai redusă (~1ms).
Diferența semantica expune atacatorului certitudinea existenței adresei de email testate,
creând un recensământ detaliat al țintelor valide.

#figure(
  caption: [TODO],
  image("assets/04-poc-04.png"),
)
