== PoC --- VULN-02: Extragere Bazelor de Date cu Parole în Clar

Această demonstrație relevă efectul stocării parolelor fără hashing în situația
compromiterii bazei de date.

*Precondiție:* Pe versiunea V1 au fost deja înregistrați utilizatori (ex. utilizatorul
creat prin exploatarea VULN-01). Se presupune că atacatorul a obținut acces la mediul
de persistență printr-o breșă secundară (simulată aici printr-un client PostgreSQL direct).

*Pași de execuție:*
Ne conectăm la instanța de PostgreSQL (`postgres`) prin utilitarul nativ `psql` și
interogăm tabelul `credentials` alături de `user` (printr-un JOIN).

```sh
docker exec -it mithras-postgres-1 psql -U user -d mithras \
  -c "SELECT u.email, c.secret FROM \"user\" u JOIN \"credential_password\" c ON c.user_pk = u.pk;"
```

*Rezultat observat:*
Interogarea returnează imediat înregistrările, expunând parola originală sub formă
de text simplu (_plaintext_) în coloana `secret`. Atacatorul nu mai trebuie să depună
niciun efort computațional pentru a obține accesul la cont, informația fiind extrasă
direct.

Dacă s-ar folosi implementarea corectă, coloana `secret` ar afișa un șir lung ce
începe cu `$argon2id$`, mascat matematic.

#figure(
  caption: [TODO],
  image("assets/02-poc-02.png"),
)
