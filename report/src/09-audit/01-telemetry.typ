== Instrumentare și Trasabilitate OpenTelemetry

Fiecare operațiune critică de securitate din cadrul IdP-ului este învelită într-un *Span*
OpenTelemetry. Această abordare permite reconstituirea exactă a lanțului de evenimente
care a condus la o decizie de autorizare sau la un eșec.

Pentru facilitarea investigațiilor (Incident Response), evenimentele de autentificare
conțin atribute standardizate, dar opace, care protejează datele sensibile (PII):
- `auth.success` (bool): indică rezultatul operațiunii.
- `auth.failure_reason` (string): motivul generic (ex. `invalid_credentials`), fără
  a expune vectorul exact.
- `user.id` (string): identificatorul public (NanoID) al utilizatorului implicat.
- `session.id` (string): identificatorul unic al sesiunii, util pentru corelarea cererilor
  ulterioare de rotație a token-urilor.

=== Demonstrarea Trasabilității

Capturile de mai jos demonstrează capacitatea platformei de telemetrie de a oferi
vizibilitate completă asupra evenimentelor de securitate.

#figure(
  caption: [Trace complet HyperDX pentru o cerere `POST /login` eșuată: ierarhia de span-uri include verificarea rate limiting, validarea OpenAPI și derivarea Argon2id (`auth.verify_password`).],
  image("assets/01-telemetry-trace.png"),
)

#figure(
  caption: [Atributele span-ului (`Column Values`) pentru cererea eșuată: `mithras.error.code` = `user.auth(invalid_credentials)`, `http.response.status_code` = `401`, politica de rate limiting activă (`login-form-per-ip`, 19 jetoane rămase).],
  image("assets/01-telemetry-col-values.png"),
)

În cazul detectării unor tipare ofensive, precum un atac de dicționar blocat de
sistemul de Rate Limiting sau o tentativă de reutilizare a unui token de refresh revocat
(*Token Replay*), sistemul emite *Span Events* specifice (`auth.lock_account_failed`,
`auth.refresh_token_replay`), semnalizând un incident de securitate iminent către
platformele SIEM.

#figure(
  caption: [Vizualizare agregată a trace-urilor în HyperDX în timpul unui atac de tip forță brută: cascadă de span-uri `POST /login` cu status `Error` la intervale de milisecunde, confirmând trasabilitatea completă a incidentului.],
  image("assets/01-telemetry-brute-force.png"),
)
