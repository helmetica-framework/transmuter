# Transmuter

**Transmute**: change from one substance into another.

The transmuter bootstraps a service helmchart for the Helmetica framework.
Once bootstrapped the service maintainer can adjust the pre-configured libraries to their liking.

## Glossary

| Term | Meaning |
| ---- | ------- |
| **Prima materia** | The raw upstream Helm chart (repository URL + version) that serves as the starting point of a transmutation. It ends up as a dependency of the reagent. |
| **Ferment** | The base chart (e.g. `oci://ghcr.io/helmetica-framework/ferment`) used as the scaffold from which the reagent is created. |
| **Azoth** | A library chart providing shared templates and helpers to reagents. |
| **Reagent** | The result of a transmutation: a valid service chart that wraps the prima materia and is ready for further development. |
| **Assay** | Non-destructive purity test of a reagent: chart validity plus CRD breaking-change detection against the latest published version of the same major. |

## Quickstart

```bash
go run . transmute \
  --name test \
  --ferment-url oci://ghcr.io/helmetica-framework/ferment:0.0.1 \
  --prima-materia-url https://charts.appcat.ch/vshnpostgresql \
  --prima-materia-version 0.8.0
```

Assay an existing reagent (validate it and check its CRD for breaking changes against the latest published version):

```bash
go run . assay \
  --path . \
  --published-url oci://ghcr.io/helmetica-framework/myreagent
```

Every flag can also be provided as an environment variable with the `TRANSMUTER_` prefix, e.g. `TRANSMUTER_FERMENT_URL`.

## Libraries

* [transmute](https://pkg.go.dev/github.com/helmetica-framework/transmuter/pkg/transmute) - Transmute a prima materia into a reagent.
* [assay](https://pkg.go.dev/github.com/helmetica-framework/transmuter/pkg/assay) - Assay a reagent and detect breaking CRD changes.
