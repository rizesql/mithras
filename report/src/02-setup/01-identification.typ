== Identificare Student și Mediu de Execuție

Conform cerințelor de predare, toate capturile de ecran incluse în acest raport conțin
prompt-ul complet al terminalului, cu username și hostname vizibile. Formatul utilizat
este:

#figure(
  caption: [Format prompt terminal - identificare student vizibilă în toate capturile],
)[
  ```
  [rizesql@mithras-vm:/Users/rizesql/fmi/anul-3/sem-2/mithras]$
  ```
]

Identificarea completă a studentului:

#figure(
  caption: [Date de identificare - student și mașină de lucru],
  table(
    columns: (auto, 1fr),
    [*Nume*], [Rizescu Iulian Ștefan],
    [*Email instituțional*], [iulian-stefan.rizescu\@s.unibuc.ro],
    [*Username terminal*], [`rizesql`],
    [*Hostname mașină*], [`mithras-vm`],
  ),
)

Mediul de execuție este o mașină virtuală NixOS gestionată prin OrbStack, care oferă
o izolare completă la nivel de sistem de operare și kernel Linux. Alegerea NixOS
garantează că întregul mediu de dezvoltare, inclusiv toolchain-ul și dependențele
sistemului, este declarat și reprodus exact așa cum este descris în configurația
proiectului, eliminând discrepanțele ("works on my machine").

#figure(
  caption: [Output `whoami` și `hostname` - confirmare identitate student pe mașina de lucru],
  image("assets/01-identification-whoami.png"),
)
