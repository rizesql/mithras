== Refresh Token și Gestiunea Sesiunilor

Sistemul de gestiune a sesiunilor în Mithras este construit pe un model hibrid de
token-uri, optimizat pentru a oferi un echilibru între performanța verificărilor locale și
capacitatea de revocare granulară a accesului. Această arhitectură separă aserțiunea de
identitate pe termen scurt de materialul de împrospătare a sesiunii pe termen lung.

=== Fluxul de Împrospătare și Rotație

Procesul de refresh este proiectat să fie atomic și auto-corectiv în cazul detectării
unor tentative de utilizare frauduloasă. Interacțiunea între componente este detaliată în
diagrama de mai jos:

#figure(
  caption: [Diagramă de Secvență - Rotația Refresh Token-ului și Detecția Anomaliilor],
  include "assets/05-session-rotation-diagram.typ",
) <fig-refresh-rotation>

=== Arhitectura Dual-Token

Mithras emite două tipuri de artefacte la fiecare autentificare reușită. Primul este
*Token-ul de Acces* (JWS), un obiect stateless cu o durată de viață extrem de scurtă (5
minute), care conține identitatea utilizatorului și rolurile acestuia semnate digital cu
EdDSA. Al doilea este *Token-ul de Refresh*, un șir de 32 de octeți generat aleatoriu, a
cărui stare este menținută în baza de date pentru o perioadă de 7 zile.

Utilizarea token-ului de refresh permite utilizatorului să obțină noi token-uri de acces
fără a reintroduce credențialele, în timp ce serverul păstrează controlul total asupra
sesiunii prin posibilitatea de a invalida oricând înregistrarea corespunzătoare din
tabelul `sessions`.

=== Rotația Token-ului de Refresh (Single-Use)

O măsură avansată de securitate implementată în versiunea securizată (*main*) este
*Rotația Token-ului de Refresh*. Fiecare token de refresh este strict de unică folosință.
În momentul în care un client utilizează un token de refresh pentru a solicita un nou
token de acces prin endpoint-ul `POST /token`, sistemul execută în mod atomic trei
operațiuni: validează și revocă token-ul curent, emite un token de refresh complet nou și
generează un nou token de acces.

Această strategie limitează semnificativ fereastra de atac în cazul interceptării unui
token. Deoarece token-ul se schimbă la fiecare utilizare, un atacator care ar fura un
token de refresh ar trebui să îl utilizeze înainte ca utilizatorul legitim să efectueze
următoarea operație de împrospătare.

Pentru a preveni invalidarea accidentală a sesiunilor legitime din cauza problemelor de
concurență (*Race Conditions*), rotația este executată într-o tranzacție de bază de date
strictă (`db.Tx`). Dacă un client web emite accidental două cereri simultane de
împrospătare cu același token, serverul garantează că doar prima cerere va reuși,
verificând numărul de rânduri afectate (`rowsAffected`) la revocarea token-ului vechi. A
doua cerere va eșua controlat, prevenind un scenariu de *Race Condition* care ar declanșa
eronat o alarmă de furt de sesiune.

=== Detecția Furtului și Revocarea Familiei de Sesiuni

Cea mai critică funcție a mecanismului de rotație este capacitatea de a detecta
compromiterea unei sesiuni. Dacă un token de refresh care a fost deja revocat (deja
utilizat) este prezentat din nou serverului, acesta este un indicator cert al unei
anomalii: fie un atacator încearcă să utilizeze un token vechi, fie utilizatorul legitim
încearcă să utilizeze un token care a fost deja utilizat anterior de un atacator.

În acest scenariu, sistemul respinge cererea curentă și invalidează automat toate
sesiunile active asociate utilizatorului respectiv. Această reacție forțează o
reautentificare completă cu credențiale primare pe toate dispozitivele, protejând contul.

#figure(
  caption: [Logica de rotație și detecție a replay-ului în `internal/auth/refresh.go`],
)[
  ```go
  func (r *Refresh) Refresh(ctx context.Context, rawToken string) (*RefreshResponse, error) {
      sess, err := r.validateSession(ctx, rawToken)

      // Detecție Replay: Dacă token-ul a fost deja revocat, invalidăm TOT
      if sess.RevokedAt != nil {
          r.revokeAllUserSessions(ctx, sess.UserPk)
          return nil, errInvalidRefreshToken("anomaly detected: token reused")
      }

      // Rotație: Revocăm token-ul vechi și inserăm unul nou în mod atomic
      return r.performRefresh(ctx, sess, now)
  }
  ```
]
