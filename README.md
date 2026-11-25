# Transmuter

**Transmute**: change from one substance into another.

The transmuter bootstraps a service Helm chart (a reagent) for the Helmetica framework from an upstream chart, then assays it and scaffolds its rituals and tests.

Documentation: https://docs.helmetica.io/transmuter/

## Quickstart

```bash
go run . transmute \
  --name test \
  --prima-materia-url https://charts.appcat.ch/vshnpostgresql \
  --prima-materia-version 0.8.0
```

## Development

```bash
make build
make test
make docs   # preview the docs
```
