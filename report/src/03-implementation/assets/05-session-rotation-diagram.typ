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

#let r_req = 0
#let r_rl = 1
#let r_lookup = 2
#let r_branch = 3
#let r_rotate = 4
#let r_revoke = 5
#let r_end = 6

#align(center)[
  #set text(font: ("CommitMono-rizesql", "Fira Code", "Consolas", "Courier New", "monospace"), size: 5.5pt)
  #diagram(
    mark-scale: 80%,

    // Entities
    node((col_bff, -0.7), [*Consumer (BFF)*], stroke: none),
    node((col_idp, -0.7), [*Mithras IdP*], stroke: none),
    node((col_infra, -0.7), [*Infrastructure*], stroke: none),

    // 1. Request
    edge((col_bff, r_req), (col_idp, r_req), "-|>", label: [POST /token (refresh_token_v1)], stroke: 1pt + fg(svelte)),

    // 2. Rate Limit
    edge((col_idp, r_rl), (col_infra, r_rl), "<->", label: [Redis: Rate Limit Check], stroke: 1pt + fg(redis)),

    // 3. DB Lookup
    edge(
      (col_idp, r_lookup),
      (col_infra, r_lookup),
      "<->",
      label: [Postgres: Get Session by SHA-256],
      stroke: 1pt + fg(pg),
    ),

    // 4. Decision Node
    node(
      (col_idp, r_branch),
      [
        *Is token valid?* \
        Active: Rotate \
        Revoked: *Family Revocation*
      ],
      stroke: 0.5pt + fg(golang),
      fill: golang.lighten(95%),
    ),

    // 5a. Success Path: Rotation
    edge(
      (col_idp, r_rotate),
      (col_infra, r_rotate),
      "-|>",
      label: [Atomic: Revoke Old + Insert New],
      stroke: 1pt + fg(pg),
    ),

    // 5b. Anomaly Path: Revoke All
    edge(
      (col_idp, r_revoke),
      (col_infra, r_revoke),
      "-|>",
      label: [Anomaly: Revoke ALL User Sessions],
      stroke: 1pt + red,
      dash: "dashed",
    ),

    // 6. Response
    edge((col_idp, r_end), (col_bff, r_end), "-|>", label: [200 (token_v2) / 401 (Anomaly)], stroke: 1pt + fg(golang)),

    // Lifelines
    edge((col_bff, 0), (col_bff, r_end), stroke: 0.5pt + gray, dash: "dashed"),
    edge((col_idp, 0), (col_idp, r_end), stroke: 0.5pt + gray, dash: "dashed"),
    edge((col_infra, 0), (col_infra, r_end), stroke: 0.5pt + gray, dash: "dashed"),
  )
]
