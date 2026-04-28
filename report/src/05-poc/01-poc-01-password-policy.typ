== PoC --- VULN-01: Înregistrare cu parolă trivială

Această demonstrație exploatează absența politicii de complexitate a parolelor din
versiunea V1.

*Precondiție:* Serverul Mithras este pornit și ascultă pe portul 8080, iar baza de
date este operațională pe branch-ul vulnerabil.

*Pași de execuție:*
Utilizăm utilitarul ofensiv dezvoltat în Go (`exploits/weak-password/main.go`) pentru a
trimite o cerere `POST` de înregistrare, alocând un șir format dintr-un singur caracter
("1") pe post de parolă.

```sh
cd exploits/weak-password
go run . -email victim@authx.com -password "1"
```

*Rezultat observat:*
Sistemul acceptă payload-ul fără a ridica erori de validare și returnează un cod HTTP `201
Created`. Contul este creat cu succes cu parola `"1"`, devenind astfel complet lipsit de
apărare împotriva unui dicționar sau atac forță brută trivial.

#figure(
  caption: [TODO],
  image("assets/01-poc-01.png"),
)
