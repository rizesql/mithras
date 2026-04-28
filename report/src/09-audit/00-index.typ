= Audit și Logging (Telemetrie)

Această secțiune documentează mecanismele de trasabilitate integrate în sistem, capabile
să detecteze și să raporteze anomaliile și tentativele de atac în timp real.

Mithras utilizează *OpenTelemetry (OTEL)* pentru a emite log-uri structurate și trace-uri
distribuite (Spans), exportate prin protocolul OTLP către un colector centralizat
(în acest caz, HyperDX/ClickHouse).

#include "01-telemetry.typ"
