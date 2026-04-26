#import "@preview/fletcher:0.5.8" as fletcher: diagram, edge, node

#align(center)[
  #set text(font: ("CommitMono-rizesql", "Fira Code", "Consolas", "Courier New", "monospace"), size: 8pt)
  #diagram(
    spacing: 0.8cm,
    mark-scale: 80%,

    node(
      (0, 0),
      [
        *1. Recunoaștere (VULN-04)* \ \
        _Enumerare email-uri valide prin_ \
        _diferențe de răspuns și timp_
      ],
      stroke: 1pt + rgb("#2F6792"),
      fill: rgb("#2F6792").lighten(90%),
      corner-radius: 4pt,
    ),

    edge((0, 0), (0, 1), "-|>", stroke: 1.5pt),

    node(
      (0, 1),
      [
        *2. Acces Inițial (VULN-01 + VULN-03)* \ \
        _Atac de dicționar rapid, neîngrădit_ \
        _pe conturi cu parole slabe_
      ],
      stroke: 1pt + rgb("#FE4236"),
      fill: rgb("#FE4236").lighten(90%),
      corner-radius: 4pt,
    ),

    edge((0, 1), (0, 2), "-|>", stroke: 1.5pt),

    node(
      (0, 2),
      [
        *3. Escaladare (Account Takeover) (VULN-06)* \ \
        _Resetare parolă cu token predictibil_ \
        _bazat pe timestamp_
      ],
      stroke: 1pt + rgb("#FF3E00"),
      fill: rgb("#FF3E00").lighten(90%),
      corner-radius: 4pt,
    ),

    edge((0, 2), (0, 3), "-|>", stroke: 1.5pt),

    node(
      (0, 3),
      [
        *4. Persistență (VULN-05)* \ \
        _Exfiltrare token sesiune prin XSS;_ \
        _token valid 24h, rezistent la logout_
      ],
      stroke: 1pt + rgb("#FE4236"),
      fill: rgb("#FE4236").lighten(90%),
      corner-radius: 4pt,
    ),

    edge((0, 3), (0, 4), "-|>", stroke: 1.5pt),

    node(
      (0, 4),
      [
        *5. Compromitere Completă (VULN-02)* \ \
        _Acces la baza de date și exfiltrarea_ \
        _parolelor în clar ale tuturor utilizatorilor_
      ],
      stroke: 1pt + rgb("#2F6792"),
      fill: rgb("#2F6792").lighten(90%),
      corner-radius: 4pt,
    ),
  )
]
