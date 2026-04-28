== PoC --- VULN-05: Furtul Sesiunii via XSS

Demonstrarea lipsei flag-ului `HttpOnly` pe cookie-ul generat de Consumer (BFF).

*Precondiție:* Mediul client (SvelteKit) rulează versiunea vulnerabilă a codului, iar
un utilizator s-a autentificat cu succes din browser-ul personal. Atacatorul a
găsit o altă vulnerabilitate XSS în aplicația frontend.

*Pași de execuție (Simulare prin DevTools):*
Odată logați în aplicație pe interfața web locală, deschidem _Developer Tools_ (F12)
în browser, navigăm la consola JavaScript și rulăm un payload de simulare a injecției:

```javascript
console.log("XSS Stolen Cookie:", document.cookie);
```

*Rezultat observat:*
Valoarea `access_token`-ului și a `refresh_token`-ului sunt vizibile în consolă, fiind
extrase integral. Dacă se folosea implementarea corectă, instrucțiunea de mai sus nu
putea citi valorile stocate cu `HttpOnly`, protejând credențialele sesiunii.

Acest token odată preluat de un actor extern poate fi inserat în propriul său client
HTTP (ex. cURL) pentru a prelua și utiliza sesiunea valabilă timp de 24 de ore, ocolind
orice acțiuni de invalidare din partea utilizatorului.

#figure(
  caption: [Demonstrația vulnerabilității VULN-05: Furtul sesiunii extrăgând cookie-ul vulnerabil lipsit de flag-ul HttpOnly din consola JavaScript.],
  image("assets/05-poc-05.png"),
)
