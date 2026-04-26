#import "@preview/fletcher:0.5.8" as fletcher: diagram, edge, node

#let svelte = rgb("#FF3E00")
#let golang = rgb("#00ADD8")
#let pg = rgb("#2F6792")
#let redis = rgb("#FE4236")
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

#let r_start = 0
#let r_rl = 0.75
#let r_lookup = 1.5
#let r_status = 2.25
#let r_verify = 3
#let r_fail = 3.75
#let r_end = 4.5
// #let r_start = 0
// #let r_rl = 1
// #let r_lookup = 2
// #let r_status = 3
// #let r_verify = 4
// #let r_fail = 5
// #let r_end = 6

#align(center)[
  #set text(font: ("CommitMono-rizesql", "Fira Code", "Consolas", "Courier New", "monospace"), size: 5.5pt)
  #diagram(
    // spacing: (3.5cm, 1cm),
    mark-scale: 80%,

    // Entities
    node((col_bff, -0.7), [*Consumer (BFF)*], stroke: none),
    node((col_idp, -0.7), [*Mithras IdP*], stroke: none),
    node((col_infra, -0.7), [*Infrastructure*], stroke: none),

    // 1. Request
    edge((col_bff, r_start), (col_idp, r_start), "-|>", label: [POST /login], stroke: 1pt + fg(svelte)),

    // 2. Rate Limit
    edge((col_idp, r_rl), (col_infra, r_rl), "<->", label: [Redis: Rate Limit Check], stroke: 1pt + fg(redis)),

    // 3. Lookup & Status
    edge((col_idp, r_lookup), (col_infra, r_lookup), "<->", label: [Postgres: Get User + Status], stroke: 1pt + fg(pg)),
    // edge((col_infra, r_lookup), (col_idp, r_lookup), "-|>", label: [Hash (or Empty)], stroke: 1pt + fg(pg)),

    // 4. Branch Logic (Decision Point)
    node(
      (col_idp, r_status),
      [
        *User exists?* \
        YES: Use DB Hash \
        NO: Use *Dummy Hash*
      ],
      stroke: 0.5pt + fg(golang),
      fill: golang.lighten(95%),
      // width: 4cm,
    ),

    // 5. Expensive Hashing (Constant Time)
    node(
      (col_idp, r_verify),
      [Argon2id Verification],
      stroke: 1pt + fg(golang),
      fill: golang.lighten(90%),
      corner-radius: 2pt,
      // width: 4cm,
    ),

    // 6. Post-auth Ops
    edge((col_idp, r_fail), (col_infra, r_fail), "-|>", label: [DB: Record Success/Fail], stroke: 1pt + fg(pg)),

    // 7. Response
    edge((col_idp, r_end), (col_bff, r_end), "-|>", label: [Uniform 200/401], stroke: 1pt + fg(golang)),

    // Lifelines
    edge((col_bff, 0), (col_bff, r_end), stroke: 0.5pt + gray, dash: "dashed"),
    edge((col_idp, 0), (col_idp, r_end), stroke: 0.5pt + gray, dash: "dashed"),
    edge((col_infra, 0), (col_infra, r_end), stroke: 0.5pt + gray, dash: "dashed"),
  )
]
