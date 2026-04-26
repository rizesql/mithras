== Lecții de Proces și Metodologie

Pe lângă implementarea tehnică, ciclul de dezvoltare *Build, Hack & Secure* a validat
o serie de principii operaționale.

=== Abordarea "Secure by Default" înaintea Sabotajului
Implementarea inițială a arhitecturii securizate (versiunea V2), urmată de
reducerea deliberată a funcțiilor de apărare pentru a crea versiunea vulnerabilă (V1), a
forțat o înțelegere profundă a necesității fiecărui control tehnic. Ordinea inversă 
(construirea unui sistem slab din start) ar fi reprezentat un risc pedagogic de asimilare 
a tiparelor arhitecturale deficitare drept comportament implicit de dezvoltare.

=== Consistența Ecosistemului de Exploatare
Scrierea scripturilor ofensive de exploatare (precum cele pentru *Brute-Force*, *User 
Enumeration* sau *Predictable Token Attack*) în același limbaj de programare cu sistemul vizat 
(Go) a permis reutilizarea logicii de concurență a goroutine-urilor și a primitivelor de timing. 
Această simetrie între stack-ul de apărare și cel de atac oferă o perspectivă unitară, 
reducând costul de context (context switch cognitiv) față de utilizarea unor unelte externe 
și demonstrând versatilitatea platformei atât pentru construirea, cât și pentru 
auditarea software-ului.