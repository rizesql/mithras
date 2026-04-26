#import "@preview/fletcher:0.5.8" as fletcher: diagram, edge, node

#let svelte = rgb("#FF3E00")
#let golang = rgb("#00ADD8")
#let pg = rgb("#2F6792")
#let redis = rgb("#FE4236")
#let browser = rgb("#555555")

#let n-style(c) = (
  fill: c.lighten(85%),
  stroke: 1pt + c.darken(30%),
  corner-radius: 2pt,
)
#let fg(c) = c.darken(30%)


// --- X-AXIS COLUMNS (Zones) ---
#let col_browser = 0
#let col_bff = 1
#let col_idp = 2
#let col_db = 3

// --- Y-AXIS LEVELS (For parallel data flows) ---
#let y_titles = -0.55
#let y_redirect = -0.3
#let y_callback = -0.15
#let y_center = 0
#let y_jwks = 0.15
#let y_refresh = 0.3

// Y-Levels to vertically stack the databases
#let y_pg = -0.25
#let y_redis = 0.25

// --- NODE COORDINATES ---
#let pos_browser = (col_browser, y_center)
#let pos_bff = (col_bff, y_center)
#let pos_idp = (col_idp, y_center)
#let pos_pg = (col_db, y_pg)
#let pos_redis = (col_db, y_redis)

#align(center)[
  #set text(font: ("CommitMono-rizesql", "Fira Code", "Consolas", "Courier New", "monospace"), size: 8pt)
  #diagram(
    spacing: (2.2cm, 2.5cm),
    mark-scale: 80%,

    // Column 0: Browser
    node((col_browser, y_titles), [*Untrusted Environment*], stroke: none),
    node(pos_browser, text(fill: fg(browser))[*Browser*], ..n-style(browser), width: 2.5cm, height: 4.5cm),

    // Column 1: BFF
    node((col_bff, y_titles), [*Confidential Client*], stroke: none),
    node(pos_bff, text(fill: fg(svelte))[*BFF*\ (SvelteKit)], ..n-style(svelte), width: 2.5cm, height: 4.5cm),

    // Column 2: IdP & Databases
    node((col_idp, y_titles), [*Root of Trust*], stroke: none),
    node(pos_idp, text(fill: fg(golang))[*Mithras IdP*], ..n-style(golang), width: 2.5cm, height: 4.5cm),

    // Column 3: Databases (Data Layer)
    node((col_db, y_titles), [*Data Layer*], stroke: none),
    node(pos_pg, text(fill: fg(pg))[PostgreSQL], ..n-style(pg), width: 2cm, height: 1.2cm),
    node(pos_redis, text(fill: fg(redis))[Redis], ..n-style(redis), width: 2cm, height: 1.2cm),

    // --- DATA FLOWS ---

    // Browser <-> BFF
    edge(pos_browser, pos_bff, "<|-|>", label: [HttpOnly Cookie\ _(Opaque)_], stroke: 1pt + fg(browser)),

    // BFF <-> IdP Parallel Flows
    edge((col_bff, y_redirect), (col_idp, y_redirect), "-|>", label: [OAuth2 Redirect], stroke: 1pt + fg(svelte)),
    edge((col_idp, y_callback), (col_bff, y_callback), "-|>", label: [OAuth2 Callback], stroke: 1pt + fg(golang)),

    edge((col_idp, y_center), (col_bff, y_center), "-|>", label: [JWS Access Token], stroke: 1pt + fg(golang)),
    edge((col_idp, y_jwks), (col_bff, y_jwks), "-|>", label: [JWKS (Startup)], stroke: 1pt + fg(golang)),

    edge((col_bff, y_refresh), (col_idp, y_refresh), "-|>", label: [Refresh Token], stroke: 1pt + fg(svelte)),

    // IdP -> Databases
    edge((col_idp, y_pg), pos_pg, "-|>", stroke: 1pt + fg(golang)),
    edge((col_idp, y_redis), pos_redis, "-|>", stroke: 1pt + fg(golang)),
  )
]
