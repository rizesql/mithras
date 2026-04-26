== Matrice de Risc

Evaluarea severității fiecărei vulnerabilități este realizată luând în considerare atât
probabilitatea de exploatare (ușurința cu care un atacator neautentificat poate executa
atacul), cât și impactul tehnic asupra sistemului.

#figure(
  caption: [Matrice de risc pentru vulnerabilitățile identificate (V1)],
  table(
    columns: (auto, auto, 1fr, auto),
    table.header[*ID*][*Probabilitate*][*Impact (CIA)*][*Severitate*],
    [VULN-01], [Ridicată], [C: Medie / I: Joasă / A: Joasă], [Medie],
    [VULN-02], [Medie], [C: Critică / I: Critică / A: Joasă], [Critică],
    [VULN-03], [Ridicată], [C: Critică / I: Medie / A: Joasă], [Critică],
    [VULN-04], [Ridicată], [C: Medie / I: Joasă / A: Joasă], [Medie],
    [VULN-05], [Medie], [C: Critică / I: Medie / A: Joasă], [Înaltă],
    [VULN-06], [Medie], [C: Critică / I: Critică / A: Medie], [Critică],
  ),
)

Probabilitatea pentru VULN-02, VULN-05 și VULN-06 este marcată ca "Medie" deoarece aceste
atacuri necesită o precondiție specifică (acces la baza de date, existența unei
vulnerabilități XSS externe, respectiv cunoașterea ferestrei de timp exacte). În contrast,
VULN-01, VULN-03 și VULN-04 au o probabilitate "Ridicată", fiind exploatabile direct
asupra endpoint-urilor publice, fără cerințe prealabile.
