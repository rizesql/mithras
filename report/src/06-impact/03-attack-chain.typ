== Lanț de Atac (Attack Chain)

Vulnerabilitățile identificate pe versiunea V1 rareori operează izolat într-un scenariu
real. Atacatorii moderni concatenează defecte de securitate minore sau medii pentru a
realiza compromiterea completă a unui sistem. Lanțul de atac cel mai probabil pentru
infrastructura AuthX urmează o progresie liniară de la recunoaștere la escaladarea
privilegiilor și exfiltrarea datelor.

#figure(
  caption: [Diagrama progresiei lanțului de atac pe versiunea V1],
  include "assets/01-attack-chain-diagram.typ",
)

Acest lanț de atac ilustrează colapsul complet al securității sistemului în starea sa
vulnerabilă, subliniind necesitatea critică a implementării corecte a principiilor de
*Hardening* prezentate în versiunea securizată (V2).
