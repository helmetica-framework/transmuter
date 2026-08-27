import "Justfile.vars.just"

export GOEXPERIMENT := "jsonv2"

_default:
    @just --list

# Build the transmuter binary, generators and checks included
build: generate fmt vet binary

# CGO is disabled here only, not globally: the image needs a static binary,
# while `just test` runs with -race, which requires cgo.
#
# Build the binary without running the generators
binary:
    @echo "GOOS=$(go env GOOS) GOARCH=$(go env GOARCH)"
    CGO_ENABLED=0 go build -o {{ bin_filename }}

# Run tests
test: generate
    go test ./... -race -coverprofile cover.out

# Run code generators
generate:
    go generate ./...

# Generate documentation
docs:
    @echo "Nothing to do yet"

# Run go fmt against code
fmt:
    go fmt ./...

# Run go vet against code
vet:
    go vet ./...

# All-in-one linting
lint: fmt vet generate docs
    @echo 'Check for uncommitted changes ...'
    git diff --exit-code

# Build the docker image
build-docker: binary
    docker build . --tag {{ GHCR_IMG }}

# Run transmuter from your host, e.g. just run transmute test oci://... https://... 0.8.0
run *args: fmt vet
    go run . {{ args }}

# Clean up the generated resources
clean:
    rm -rf dist/ cover.out {{ bin_filename }} || true
