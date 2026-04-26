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
#let r_bff = 1.5
#let r_hash = 2.25
#let r_lookup = 3
#let r_anomaly = 3.75
#let r_revoke = 4.5
#let r_end = 5.25

// #let r_start = 0
// #let r_rl = 1
// #let r_bff = 2
// #let r_hash = 3
// #let r_lookup = 4
// #let r_anomaly = 5
// #let r_revoke = 6
// #let r_end = 7

#align(center)[
  #set text(font: ("CommitMono-rizesql", "Fira Code", "Consolas", "Courier New", "monospace"), size: 5.5pt)
  #diagram(
    mark-scale: 80%,

    // Entities
    node((col_bff, -0.7), [*Consumer (BFF)*], stroke: none),
    node((col_idp, -0.7), [*Mithras IdP*], stroke: none),
    node((col_infra, -0.7), [*Infrastructure*], stroke: none),

    // 1. Request
    edge((col_bff, r_start), (col_idp, r_start), "-|>", label: [POST /logout], stroke: 1pt + fg(svelte)),

    // 2. Rate Limit (Redis)
    edge((col_idp, r_rl), (col_infra, r_rl), "<->", label: [Redis: Rate Limit Check], stroke: 1pt + fg(redis)),


    // 3. BFF Cookie Cleanup
    node(
      (col_bff, r_bff),
      [Șterge Cookie HttpOnly],
      stroke: 0.5pt + fg(svelte),
      fill: svelte.lighten(95%),
      corner-radius: 2pt,
    ),

    // 4. IdP Token Hash
    node(
      (col_idp, r_hash),
      [SHA-256(RefreshToken)],
      stroke: 0.5pt + fg(golang),
      fill: golang.lighten(95%),
      corner-radius: 2pt,
    ),

    // 5. DB Lookup
    edge((col_idp, r_lookup), (col_infra, r_lookup), "<->", label: [Postgres: Get Session], stroke: 1pt + fg(pg)),

    // 6. Anomaly Detection
    node(
      (col_idp, r_anomaly),
      [
        *Already revoked?* \
        YES: Threat Alert \
        NO: Standard Logout
      ],
      stroke: 0.5pt + fg(golang),
      fill: golang.lighten(95%),
    ),

    // 7. Force Logout All
    edge(
      (col_idp, r_revoke + 0.25),
      (col_infra, r_revoke + 0.25),
      "-|>",
      label: [Postgres: RevokeUserSessions],
      stroke: 1pt + fg(pg),
    ),

    // 8. Response
    edge((col_idp, r_end), (col_bff, r_end), "-|>", label: [200 OK / 401], stroke: 1pt + fg(golang)),

    // Lifelines
    edge((col_bff, 0), (col_bff, r_end), stroke: 0.5pt + gray, dash: "dashed"),
    edge((col_idp, 0), (col_idp, r_end), stroke: 0.5pt + gray, dash: "dashed"),
    edge((col_infra, 0), (col_infra, r_end), stroke: 0.5pt + gray, dash: "dashed"),
  )
]
