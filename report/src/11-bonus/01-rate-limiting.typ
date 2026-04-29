== Rate Limiting Avansat (Custom Lua Token Bucket)

Mecanismul de protecție împotriva atacurilor de forță brută (VULN-03) și epuizare a
resurselor (DoS) a fost realizat printr-o implementare originală, de la zero, a
algoritmului *Token Bucket* în limbajul Go și Lua, fără a depinde de biblioteci externe de
rate limiting.

Logica atomică este delegată direct motorului Redis printr-un script Lua, garantând că
evaluarea numărului curent de jetoane, decizia de respingere și reîncărcarea capacității
se realizează într-un singur ciclu non-blocant, perfect imun la stările de cursă (Race
Conditions) sub trafic extrem.

Această implementare permite declararea unor politici compuse independente. De exemplu,
endpoint-ul de login beneficiază simultan de:
- O politică *global-per-ip* (limită mare, reîncărcare rapidă) pentru a bloca inundarea
  generală a rețelei (Volumetric DoS).
- O politică *strict-per-account* pentru blocarea atacurilor de forță brută direcționate
  (Dictionary Attacks).

În plus, motorul de rate limiting acceptă politici configurabile cu strategii de degradare
gracioasă de tip *Fail-Open* (în cazul unui sistem de logare care pică) sau *Fail-Closed*
(esențial pentru mecanisme de autentificare).

Pentru a asigura atomicitatea și performanța fără a aglomera codul aplicației, calculul
jetoanelor este realizat direct în Redis. Nucleul logicii de calcul, extras din scriptul
Lua, este următorul:

```lua
local data = redis.call("HMGET", key, "tokens", "ts")
-- parsarea tokens, ts din data...

local delta = math.max(0, now - ts)
local refill = delta * rate
tokens = math.min(capacity, tokens + refill)

if tokens >= 1 then
    tokens = tokens - 1
    allowed = 1
else
    allowed = 0
    retry_after = math.ceil((1 - tokens) / rate)
end

redis.call("HSET", key, "tokens", tokens, "ts", now)

local ttl = math.ceil(capacity / rate)
redis.call("PEXPIRE", key, ttl)

local remaining = math.floor(tokens)
return {allowed, retry_after, remaining}
```

Prin rularea acestui script, Redis actualizează atomic numărul de jetoane disponibile și
returnează decizia (`allowed`), împreună cu timpul necesar până la următoarea încercare
validă (`retry_after`). Această delegare elimină overhead-ul de rețea cauzat de interogări
repetate și previne complet condițiile de cursă (Race Conditions) la nivelul stocării.
