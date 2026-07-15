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
| **Ritual** | A packaged `Definition` manifest (`rituals.helmetica.io/v1`) in a reagent describing a single or scheduled operational action (e.g. restart, maintenance). Executed at runtime by a separate controller via `Action` CRs; the transmuter only scaffolds and assays them. |

## Quickstart

```bash
go run . transmute \
  --name test \
  --prima-materia-url https://charts.appcat.ch/vshnpostgresql \
  --prima-materia-version 0.8.0
```

`--ferment-url` is optional and defaults to `oci://ghcr.io/helmetica-framework/ferment`.
A ferment URL without a tag resolves to the latest available version.

Assay an existing reagent (validate it and check its CRD for breaking changes against the latest published version):

```bash
go run . assay \
  --path . \
  --published-url oci://ghcr.io/helmetica-framework/myreagent
```

Add a skeleton ritual to an existing reagent:

```bash
go run . ritual add \
  --path . \
  --name restart
```

Assay validates every ritual `Definition` found in the rendered reagent.

Every flag can also be provided as an environment variable with the `TRANSMUTER_` prefix, e.g. `TRANSMUTER_FERMENT_URL`.

## Libraries

* [transmute](https://pkg.go.dev/github.com/helmetica-framework/transmuter/pkg/transmute) - Transmute a prima materia into a reagent.
* [assay](https://pkg.go.dev/github.com/helmetica-framework/transmuter/pkg/assay) - Assay a reagent and detect breaking CRD changes.
* [ritual](https://pkg.go.dev/github.com/helmetica-framework/transmuter/pkg/ritual) - Scaffold ritual Definitions into a reagent.
