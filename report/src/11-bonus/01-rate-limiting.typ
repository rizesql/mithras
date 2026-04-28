== Rate Limiting Avansat (Custom Lua Token Bucket)

Mecanismul de protecție împotriva atacurilor de forță brută (VULN-03) și epuizare
a resurselor (DoS) a fost realizat printr-o implementare originală, de la zero, a
algoritmului *Token Bucket* în limbajul Go și Lua, fără a depinde de biblioteci
externe de rate limiting.

Logica atomică este delegată direct motorului Redis printr-un script Lua, garantând
că evaluarea numărului curent de jetoane, decizia de respingere și reîncărcarea
capacității se realizează într-un singur ciclu non-blocant, perfect imun la
stările de cursă (Race Conditions) sub trafic extrem.

Această implementare permite declararea unor politici compuse independente. De
exemplu, endpoint-ul de login beneficiază simultan de:
- O politică *global-per-ip* (limită mare, reîncărcare rapidă) pentru a bloca
  inundarea generală a rețelei (Volumetric DoS).
- O politică *strict-per-account* pentru blocarea atacurilor de forță brută
  direcționate (Dictionary Attacks).

În plus, motorul de rate limiting acceptă politici configurabile cu strategii de
degradare gracioasă de tip *Fail-Open* (în cazul unui sistem de logare care pică) 
sau *Fail-Closed* (esențial pentru mecanisme de autentificare).

_[Aici va fi introdusă CAPTURA cu codul Lua din `redis.go`]_
