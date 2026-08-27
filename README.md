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
| **Touchstone** | An end-to-end [chainsaw](https://kyverno.github.io/chainsaw/) test of a reagent against a running athanor cluster. Lives in `test/touchstone/<name>/`; the transmuter only scaffolds them, `just touchstone` runs them. |
| **Computed values** | What a reagent's `cel:` expressions evaluate to for a given claim and its values. Helm cannot evaluate them, so `transmuter values` writes them to a file the unit tests read. |
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

The reagent's `Chart.yaml` is the ferment's, under the reagent's name and with the prima
materia added to its dependencies. Whatever else the ferment declares, the azoth version
it pins included, carries over, so the ferment stays the one place a reagent's metadata is
decided. The reagent needs `helm dependency build` before it renders, which `just test`,
`just build` and the touchstone all do.

`--path` is optional on every command that takes it and defaults to the current directory, so the
commands below are meant to be run from inside the reagent.

Assay an existing reagent (validate it and check its CRD for breaking changes against the latest published version):

```bash
go run . assay \
  --published-url oci://ghcr.io/helmetica-framework/myreagent
```

Add a skeleton ritual to an existing reagent:

```bash
go run . ritual add --name restart
```

Assay validates every ritual `Definition` found in the rendered reagent.

Add a skeleton touchstone (chainsaw test) to an existing reagent:

```bash
go run . touchstone add --name install
```

This writes a chainsaw template test to `test/touchstone/<name>/`.
The template will have various variables pre-defined that help testing reagents.
Take a look at the template to find out where to add assertions for the test.

Generate the values a reagent's `cel:` expressions compute:

```bash
go run . values --name instance
```

This will create a values file containing computed values for all the cel expression in the reagent.
Passing this values file can then be used together with vanilla helm commands.


Every flag can also be provided as an environment variable with the `TRANSMUTER_` prefix, e.g. `TRANSMUTER_FERMENT_URL`.

## Libraries

* [transmute](https://pkg.go.dev/github.com/helmetica-framework/transmuter/pkg/transmute) - Transmute a prima materia into a reagent.
* [assay](https://pkg.go.dev/github.com/helmetica-framework/transmuter/pkg/assay) - Assay a reagent and detect breaking CRD changes.
* [ritual](https://pkg.go.dev/github.com/helmetica-framework/transmuter/pkg/ritual) - Scaffold ritual Definitions into a reagent.
* [touchstone](https://pkg.go.dev/github.com/helmetica-framework/transmuter/pkg/touchstone) - Scaffold chainsaw tests into a reagent.
* [values](https://pkg.go.dev/github.com/helmetica-framework/transmuter/pkg/values) - Generate the values a reagent's cel: expressions compute.
