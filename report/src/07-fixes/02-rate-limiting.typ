== Remediere VULN-03: Rate Limiting (Redis Token Bucket)

Lipsa restricțiilor pentru atacurile de tip brute-force (CWE-307) a fost neutralizată
prin reintroducerea și configurarea middleware-ului de Rate Limiting pe întregul strat
de autentificare.

=== Implementarea Backend-ului Redis

Omiterea deliberată a limitatorului de pe branch-ul vulnerabil a fost corectată pe
branch-ul securizat prin invocarea unui backend real bazat pe Redis. Sistemul utilizează
algoritmul *Token Bucket* implementat printr-un script Lua. Atomicitatea execuției la
nivelul serverului Redis asigură că decrementarea "jetoanelor" și verificarea limitelor
nu sunt predispuse la stări de cursă (*Race Conditions*), chiar și sub un trafic extrem
de concurent creat de un atacator.

=== Politici Compuse și Fail-Closed

Endpoint-urile critice, precum `/login` și `/register`, sunt protejate prin politici
configurabile. Sistemul impune o limită globală per adresă IP (ex. 1000 de cereri pe
minut) pentru protecția împotriva abuzurilor DDoS, suplimentată de un mecanism de *Account
Lockout* (blocarea temporală a contului după 5 încercări eșuate consecutive).

O decizie importantă de design integrată în middleware este adoptarea strategiei
*Fail-Closed*: dacă serverul Redis devine indisponibil, cererile de autentificare sunt
respinse implicit cu un cod de eroare severă (HTTP 500). Această alegere previne lăsarea
sistemului neprotejat la forța brută sub pretextul degradării infrastructurii.
