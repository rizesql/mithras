#import "@preview/fletcher:0.5.8" as fletcher: diagram, edge, node

#let svelte = rgb("#FF3E00")
#let golang = rgb("#00ADD8")
#let redis = rgb("#FE4236")
#let pg = rgb("#2F6792")
#let fg(c) = c.darken(30%)

#let n-style(c) = (
  fill: c.lighten(85%),
  stroke: 1pt + c.darken(30%),
  corner-radius: 2pt,
)

// --- GRID ---
#let col_bff = 0
#let col_idp = 1
#let col_infra = 2

// Rows: Sequence of operations
#let r_start = 0
#let r_rl = 0.75
#let r_val = 1.5
#let r_hash = 2.25
#let r_tx = 3
#let r_db = 3.75
#let r_end = 4.5
// #let r_start = 0
// #let r_rl = 1
// #let r_val = 2
// #let r_hash = 3
// #let r_tx = 4
// #let r_db = 5
// #let r_end = 6

#align(center)[
  #set text(font: ("CommitMono-rizesql", "Fira Code", "Consolas", "Courier New", "monospace"), size: 5.5pt)
  #diagram(
    // spacing: (2.5cm, 0.8cm),
    mark-scale: 80%,

    // Lifelines / Entities
    node((col_bff, -0.7), [*Consumer (BFF)*], stroke: none),
    node((col_idp, -0.7), [*Mithras IdP*], stroke: none),
    node((col_infra, -0.7), [*Infrastructure*], stroke: none),

    // 1. Request
    edge((col_bff, r_start), (col_idp, r_start), "-|>", label: [POST /register], stroke: 1pt + fg(svelte)),

    edge((col_idp, r_rl), (col_infra, r_rl), "<->", label: [Redis: Rate Limit Check], stroke: 1pt + fg(redis)),

    // 2. Validation (Internal IdP)
    node(
      (col_idp, r_val),
      [
        Validare \
        (Email, Password)
      ],
      stroke: 0.5pt + fg(golang),
      fill: golang.lighten(95%),
      corner-radius: 2pt,
      // width: 2.3cm,
    ),

    // 3. Hashing (Internal IdP)
    node(
      (col_idp, r_hash),
      [Argon2id Hashing],
      stroke: 0.5pt + fg(golang),
      fill: golang.lighten(95%),
      corner-radius: 2pt,
      // width: 2.3cm,
    ),

    // 4. DB Transaction
    edge((col_idp, r_tx), (col_infra, r_tx), "-|>", label: [BEGIN TRANSACTION], stroke: 1pt + fg(golang)),

    // 5. DB Ops (Grouped)
    node(
      (col_infra, r_db),
      [
        INSERT users \
        INSERT credentials \
        INSERT history
      ],
      ..n-style(pg),
      // width: 3.5cm,
    ),
    edge((col_idp, r_db), (col_infra, r_db), "<->", stroke: 1pt + fg(golang)),

    // 6. Commit
    edge((col_idp, r_tx + 0.75), (col_infra, r_tx + 0.75), "-|>", label: [COMMIT], stroke: 1pt + fg(golang)),

    // 7. Response
    edge(
      (col_idp, r_end),
      (col_bff, r_end),
      "-|>",
      label: [201 Created (JWS + Refresh)],
      stroke: 1pt + fg(golang),
    ),

    // Vertical lines (Lifelines)
    edge((col_bff, 0), (col_bff, r_end), stroke: 0.5pt + gray, dash: "dashed"),
    edge((col_idp, 0), (col_idp, r_end), stroke: 0.5pt + gray, dash: "dashed"),
    edge((col_infra, 0), (col_infra, r_end), stroke: 0.5pt + gray, dash: "dashed"),
  )
]
