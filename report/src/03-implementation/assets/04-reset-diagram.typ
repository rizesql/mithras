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
#let col_user = 0
#let col_idp = 1
#let col_infra = 2

#let r_req = 0
#let r_rl1 = 0.75
#let r_gen = 1.5
#let r_insert = 2.25
#let r_mail = 3
#let r_reset = 3.75
#let r_rl2 = 4.5
#let r_val = 5.25
#let r_tx = 6
#let r_db = 6.75
#let r_end = 7.5
// #let r_req = 0
// #let r_rl1 = 1
// #let r_gen = 2
// #let r_insert = 3
// #let r_mail = 4
// #let r_reset = 5
// #let r_rl2 = 6
// #let r_val = 7
// #let r_tx = 8
// #let r_db = 9
// #let r_end = 10

#align(center)[
  #set text(font: ("CommitMono-rizesql", "Fira Code", "Consolas", "Courier New", "monospace"), size: 5.5pt)
  #diagram(
    mark-scale: 80%,

    // Entities
    node((col_user, -0.7), [*User / BFF*], stroke: none),
    node((col_idp, -0.7), [*Mithras IdP*], stroke: none),
    node((col_infra, -0.7), [*Infrastructure*], stroke: none),

    // Phase 1: Request
    edge((col_user, r_req), (col_idp, r_req), "-|>", label: [POST /forgot-password], stroke: 1pt + fg(svelte)),

    edge((col_idp, r_rl1), (col_infra, r_rl1), "<->", label: [Redis: Rate Limit Check], stroke: 1pt + fg(redis)),

    node(
      (col_idp, r_gen),
      [
        Generare Token \
        `SHA-256(token)`
      ],
      stroke: 0.5pt + fg(golang),
      fill: golang.lighten(95%),
    ),

    edge(
      (col_idp, r_insert),
      (col_infra, r_insert),
      "-|>",
      label: [Postgres: Insert Token (Hash)],
      stroke: 1pt + fg(pg),
    ),

    node(
      (col_user, r_mail - 0.25),
      [
        User accesează \
        link-ul
      ],
      stroke: (thickness: 0.5pt, paint: gray, dash: "dashed"),
      fill: gray.lighten(95%),
    ),

    // Phase 2: Reset
    edge(
      (col_user, r_reset),
      (col_idp, r_reset),
      "-|>",
      label: [POST /reset-password (Token + NewPwd)],
      stroke: 1pt + fg(svelte),
    ),

    edge((col_idp, r_rl2), (col_infra, r_rl2), "<->", label: [Redis: Rate Limit Check], stroke: 1pt + fg(redis)),

    node(
      (col_idp, r_val),
      [
        Validare Token (Hash Check)\
        Check Pwd History (Last 5)
      ],
      stroke: 0.5pt + fg(golang),
      fill: golang.lighten(95%),
    ),

    edge((col_idp, r_tx), (col_infra, r_tx), "-|>", label: [BEGIN TRANSACTION], stroke: 1pt + fg(golang)),

    node(
      (col_infra, r_db),
      [
        UPDATE credentials \
        MARK token used \
        INVALIDATE siblings \
        REVOKE all sessions
      ],
      ..n-style(pg),
    ),
    edge((col_idp, r_db), (col_infra, r_db), "<->", stroke: 1pt + fg(golang)),

    edge((col_idp, r_tx + 0.75), (col_infra, r_tx + 0.75), "-|>", label: [COMMIT], stroke: 1pt + fg(golang)),

    edge((col_idp, r_end), (col_user, r_end), "-|>", label: [200 OK], stroke: 1pt + fg(golang)),

    // Lifelines
    edge((col_user, 0), (col_user, r_end), stroke: 0.5pt + gray, dash: "dashed"),
    edge((col_idp, 0), (col_idp, r_end), stroke: 0.5pt + gray, dash: "dashed"),
    edge((col_infra, 0), (col_infra, r_end), stroke: 0.5pt + gray, dash: "dashed"),
  )
]
