== PoC --- VULN-03: Brute Force fără Blocare Temporară

Demonstrăm transformarea forței brute într-un mecanism de pătrundere complet viabil pe
baza absenței rate limiting-ului.

*Precondiție:* Există contul victimei `victim@authx.com` cu parola `1` generat în VULN-01.

*Pași de execuție:*
Rulăm script-ul specializat de atac `exploits/brute-force/main.go`, care este configurat
să parcurgă secvențial o listă (dicționar) de 40 de parole posibile. Atacul este orientat
către endpoint-ul de login, fără pauze între cereri.

```sh
cd exploits/brute-force
go run . -email victim@authx.com
```

*Rezultat observat:*
Output-ul script-ului demonstrează iterarea fluentă prin dicționar, marcând
zeci de răspunsuri consecutive `401 Unauthorized`. Sistemul IdP nu blochează executarea.
În momentul în care dicționarul testează valoarea "1", se returnează `200 OK`, iar
script-ul raportează "Password found". Efortul atacatorului a necesitat zero pauze sau
modificări de proxy/IP.


#figure(
  caption: [Demonstrația vulnerabilității VULN-03: Atac de tip forță brută reușit prin testarea secvențială fără blocare temporară.],
  image("assets/03-poc-03.png"),
)
