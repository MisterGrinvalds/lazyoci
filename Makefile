.PHONY: build run test clean install install-completions uninstall lint fmt help \
	registry-up registry-down registry-logs registry-push-test \
	push-image push-helm push-sbom-spdx push-sbom-cyclonedx \
	push-signature push-attestation push-wasm registry-push-all \
	test-registry test-config test-cache test-pull test-artifacts test-build test-all \
	docs-install docs-dev docs-build docs-serve \
	release-token release-test release-dry-run \
	build-local build-docr \
	test-docr-helm test-docr-artifact test-docr-docker test-docr-multi test-docr-all

# Build variables
BINARY_NAME := lazyoci
BUILD_DIR := bin
MAIN_PATH := ./cmd/lazyoci
PREFIX ?= /usr/local

# Go variables
GOFLAGS := -ldflags="-s -w"

# Registry variables
REGISTRY := localhost:5050
ORAS_FLAGS := --insecure

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

## build: Build the binary
build:
	@mkdir -p $(BUILD_DIR)
	go build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

## run: Run the application
run: build
	$(BUILD_DIR)/$(BINARY_NAME)

## test: Run tests
test:
	go test -v ./...

## test-coverage: Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

## lint: Run linter
lint:
	golangci-lint run

## fmt: Format code
fmt:
	go fmt ./...
	goimports -w .

## clean: Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -rf completions
	rm -f coverage.out coverage.html

## install: Install the binary to PREFIX/bin (default /usr/local/bin)
install: build
	install -d $(PREFIX)/bin
	install -m 755 $(BUILD_DIR)/$(BINARY_NAME) $(PREFIX)/bin/$(BINARY_NAME)

## install-completions: Generate and install shell completions
install-completions: build
	@echo "Generating shell completions..."
	@mkdir -p completions
	$(BUILD_DIR)/$(BINARY_NAME) completion bash > completions/$(BINARY_NAME).bash
	$(BUILD_DIR)/$(BINARY_NAME) completion zsh > completions/_$(BINARY_NAME)
	$(BUILD_DIR)/$(BINARY_NAME) completion fish > completions/$(BINARY_NAME).fish
	install -d $(PREFIX)/share/bash-completion/completions
	install -m 644 completions/$(BINARY_NAME).bash $(PREFIX)/share/bash-completion/completions/$(BINARY_NAME)
	install -d $(PREFIX)/share/zsh/site-functions
	install -m 644 completions/_$(BINARY_NAME) $(PREFIX)/share/zsh/site-functions/_$(BINARY_NAME)
	install -d $(PREFIX)/share/fish/completions
	install -m 644 completions/$(BINARY_NAME).fish $(PREFIX)/share/fish/completions/$(BINARY_NAME).fish
	@echo "Shell completions installed to $(PREFIX)/share/"

## uninstall: Remove binary and completions from PREFIX
uninstall:
	rm -f $(PREFIX)/bin/$(BINARY_NAME)
	rm -f $(PREFIX)/share/bash-completion/completions/$(BINARY_NAME)
	rm -f $(PREFIX)/share/zsh/site-functions/_$(BINARY_NAME)
	rm -f $(PREFIX)/share/fish/completions/$(BINARY_NAME).fish

## deps: Download dependencies
deps:
	go mod download
	go mod tidy

## dev: Run with hot reload (requires air)
dev:
	air -c .air.toml

# ---------------------------------------------------------------------------
# Local OCI Registry (development)
# ---------------------------------------------------------------------------

## registry-up: Start local OCI registry on localhost:5050
registry-up:
	docker compose -f docker-compose.dev.yml up -d
	@echo "Registry running at $(REGISTRY)"
	@echo "Add to lazyoci: lazyoci registry add $(REGISTRY) --insecure"

## registry-down: Stop and remove local OCI registry (including data)
registry-down:
	docker compose -f docker-compose.dev.yml down -v

## registry-logs: Tail local registry logs
registry-logs:
	docker compose -f docker-compose.dev.yml logs -f registry

## registry-push-test: Push a sample generic OCI artifact
registry-push-test:
	@command -v oras >/dev/null 2>&1 || { echo "oras CLI required: brew install oras or https://oras.land/docs/installation"; exit 1; }
	@echo "Pushing sample artifact to $(REGISTRY)/test/hello:v1 ..."
	echo '{"message": "hello from lazyoci"}' > lazyoci-test-artifact.json
	oras push $(ORAS_FLAGS) $(REGISTRY)/test/hello:v1 \
		lazyoci-test-artifact.json:application/vnd.lazyoci.test.v1+json
	rm -f lazyoci-test-artifact.json
	@echo "Verifying via catalog API ..."
	@curl -sf http://$(REGISTRY)/v2/_catalog | python3 -m json.tool
	@echo "Done. Add registry and browse:"
	@echo "  ./bin/lazyoci registry add $(REGISTRY) --insecure"
	@echo "  ./bin/lazyoci"

# ---------------------------------------------------------------------------
# OCI Artifact Type Fixtures (push to local registry)
# ---------------------------------------------------------------------------
# Each target pushes a realistic artifact with correct OCI media types so
# that lazyoci's type detection can identify them in the TUI.
# Requires: oras CLI (brew install oras) and a running local registry.
# Fixture payloads live in testdata/fixtures/.

FIXTURES_DIR := testdata/fixtures

define check_oras
	@command -v oras >/dev/null 2>&1 || { echo "oras CLI required: brew install oras or https://oras.land/docs/installation"; exit 1; }
endef

## push-image: Push a minimal container image (OCI layout)
push-image:
	$(call check_oras)
	@echo "==> Pushing container image to $(REGISTRY)/test/myapp:v1.0.0 ..."
	@mkdir -p .tmp-fixtures
	@echo '{"architecture":"amd64","os":"linux","rootfs":{"type":"layers","diff_ids":["sha256:0000000000000000000000000000000000000000000000000000000000000000"]},"config":{}}' > .tmp-fixtures/config.json
	@printf '#!/bin/sh\necho "hello from lazyoci test image"\n' > .tmp-fixtures/entrypoint.sh
	oras push $(ORAS_FLAGS) $(REGISTRY)/test/myapp:v1.0.0 \
		--config .tmp-fixtures/config.json:application/vnd.oci.image.config.v1+json \
		.tmp-fixtures/entrypoint.sh:application/vnd.oci.image.layer.v1.tar+gzip
	@rm -rf .tmp-fixtures
	@echo "    Container Image pushed successfully"

## push-helm: Push a Helm chart artifact
push-helm:
	$(call check_oras)
	@echo "==> Pushing Helm chart to $(REGISTRY)/test/mychart:0.1.0 ..."
	@mkdir -p .tmp-fixtures/mychart
	@echo '{"name":"mychart","version":"0.1.0","description":"A test Helm chart for lazyoci","apiVersion":"v2","type":"application"}' > .tmp-fixtures/chart-config.json
	@printf 'apiVersion: v2\nname: mychart\nversion: 0.1.0\ndescription: A test Helm chart for lazyoci\n' > .tmp-fixtures/mychart/Chart.yaml
	@printf 'apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: test-config\ndata:\n  greeting: hello\n' > .tmp-fixtures/mychart/configmap.yaml
	@tar czf .tmp-fixtures/mychart-0.1.0.tgz -C .tmp-fixtures mychart
	oras push $(ORAS_FLAGS) $(REGISTRY)/test/mychart:0.1.0 \
		--config .tmp-fixtures/chart-config.json:application/vnd.cncf.helm.config.v1+json \
		.tmp-fixtures/mychart-0.1.0.tgz:application/vnd.cncf.helm.chart.content.v1.tar+gzip
	@rm -rf .tmp-fixtures
	@echo "    Helm Chart pushed successfully"

## push-sbom-spdx: Push an SPDX SBOM artifact
push-sbom-spdx:
	$(call check_oras)
	@echo "==> Pushing SPDX SBOM to $(REGISTRY)/test/myapp-sbom:spdx-v1 ..."
	oras push $(ORAS_FLAGS) $(REGISTRY)/test/myapp-sbom:spdx-v1 \
		$(FIXTURES_DIR)/sbom-spdx.json:application/spdx+json
	@echo "    SPDX SBOM pushed successfully"

## push-sbom-cyclonedx: Push a CycloneDX SBOM artifact
push-sbom-cyclonedx:
	$(call check_oras)
	@echo "==> Pushing CycloneDX SBOM to $(REGISTRY)/test/myapp-sbom:cyclonedx-v1 ..."
	oras push $(ORAS_FLAGS) $(REGISTRY)/test/myapp-sbom:cyclonedx-v1 \
		$(FIXTURES_DIR)/sbom-cyclonedx.json:application/vnd.cyclonedx+json
	@echo "    CycloneDX SBOM pushed successfully"

## push-signature: Push a cosign signature artifact
push-signature:
	$(call check_oras)
	@echo "==> Pushing cosign signature to $(REGISTRY)/test/myapp-sig:sha256-abc123 ..."
	oras push $(ORAS_FLAGS) $(REGISTRY)/test/myapp-sig:sha256-abc123 \
		$(FIXTURES_DIR)/cosign-sig.json:application/vnd.dev.cosign.simplesigning.v1+json
	@echo "    Cosign Signature pushed successfully"

## push-attestation: Push an in-toto attestation artifact
push-attestation:
	$(call check_oras)
	@echo "==> Pushing in-toto attestation to $(REGISTRY)/test/myapp-att:v1 ..."
	oras push $(ORAS_FLAGS) $(REGISTRY)/test/myapp-att:v1 \
		$(FIXTURES_DIR)/attestation-intoto.json:application/vnd.in-toto+json
	@echo "    In-toto Attestation pushed successfully"

## push-wasm: Push a WebAssembly module artifact
push-wasm:
	$(call check_oras)
	@echo "==> Pushing WASM module to $(REGISTRY)/test/myapp-wasm:v1 ..."
	@mkdir -p .tmp-fixtures
	@printf '\x00asm\x01\x00\x00\x00' > .tmp-fixtures/module.wasm
	oras push $(ORAS_FLAGS) $(REGISTRY)/test/myapp-wasm:v1 \
		.tmp-fixtures/module.wasm:application/vnd.wasm.content.layer.v1+wasm
	@rm -rf .tmp-fixtures
	@echo "    WebAssembly Module pushed successfully"

## registry-push-all: Push all OCI artifact type fixtures to local registry
registry-push-all: push-image push-helm push-sbom-spdx push-sbom-cyclonedx push-signature push-attestation push-wasm
	@echo ""
	@echo "All artifact types pushed to $(REGISTRY):"
	@echo "  Container Image     $(REGISTRY)/test/myapp:v1.0.0"
	@echo "  Helm Chart          $(REGISTRY)/test/mychart:0.1.0"
	@echo "  SBOM (SPDX)         $(REGISTRY)/test/myapp-sbom:spdx-v1"
	@echo "  SBOM (CycloneDX)    $(REGISTRY)/test/myapp-sbom:cyclonedx-v1"
	@echo "  Signature (Cosign)  $(REGISTRY)/test/myapp-sig:sha256-abc123"
	@echo "  Attestation (SLSA)  $(REGISTRY)/test/myapp-att:v1"
	@echo "  WebAssembly         $(REGISTRY)/test/myapp-wasm:v1"
	@echo ""
	@echo "Verify: curl -s http://$(REGISTRY)/v2/_catalog | python3 -m json.tool"
	@echo "Browse: ./bin/lazyoci"

# ---------------------------------------------------------------------------
# Component Tests
# ---------------------------------------------------------------------------
# Run focused tests for individual packages. Use "make test" to run all.

## test-registry: Test registry client (type detection, sorting, parsing, credentials)
test-registry:
	go test -v ./pkg/registry/...

## test-config: Test configuration (paths, artifact dir priority, registry CRUD)
test-config:
	go test -v ./pkg/config/...

## test-cache: Test metadata cache (set/get, TTL expiry, key sanitization)
test-cache:
	go test -v ./pkg/cache/...

## test-pull: Test pull logic (reference parsing, type detection, directory mapping)
test-pull:
	go test -v ./pkg/pull/...

## test-artifacts: Test artifact handlers (dispatch, actions, details)
test-artifacts:
	go test -v ./pkg/artifacts/...

## test-build: Test build system (config parsing, template rendering, validation)
test-build:
	go test -v ./pkg/build/...

## test-all: Run all unit tests with race detection
test-all:
	go test -v -race -count=1 ./...

# ---------------------------------------------------------------------------
# Example Builds
# ---------------------------------------------------------------------------

DOCR := registry.digitalocean.com/greenforests
DOCR_TAG ?= v0.1.0-test

## build-local: Dry-run build examples against local registry (localhost:5050)
build-local: build
	LAZYOCI_REGISTRY=localhost:5050 $(BUILD_DIR)/$(BINARY_NAME) build \
		--file examples/multi/.lazy --tag $(DOCR_TAG) --insecure --dry-run

## build-docr: Dry-run build examples against DigitalOcean registry
build-docr: build
	LAZYOCI_REGISTRY=$(DOCR) $(BUILD_DIR)/$(BINARY_NAME) build \
		--file examples/multi/.lazy --tag $(DOCR_TAG) --dry-run

## test-docr-helm: Build and push Helm chart to DigitalOcean registry
test-docr-helm: build
	LAZYOCI_REGISTRY=$(DOCR) $(BUILD_DIR)/$(BINARY_NAME) build \
		--file examples/helm/.lazy --tag $(DOCR_TAG)

## test-docr-artifact: Build and push generic artifact to DigitalOcean registry
test-docr-artifact: build
	LAZYOCI_REGISTRY=$(DOCR) $(BUILD_DIR)/$(BINARY_NAME) build \
		--file examples/artifact/.lazy --tag $(DOCR_TAG)

## test-docr-docker: Build and push Docker daemon image to DigitalOcean registry
test-docr-docker: build
	@docker pull nginx:alpine 2>/dev/null || true
	LAZYOCI_REGISTRY=$(DOCR) $(BUILD_DIR)/$(BINARY_NAME) build \
		--file examples/docker/.lazy --tag $(DOCR_TAG)

## test-docr-multi: Build and push all multi-example artifacts to DigitalOcean registry
test-docr-multi: build
	LAZYOCI_REGISTRY=$(DOCR) $(BUILD_DIR)/$(BINARY_NAME) build \
		--file examples/multi/.lazy --tag $(DOCR_TAG)

## test-docr-all: Build and push all examples to DigitalOcean registry
test-docr-all: test-docr-helm test-docr-artifact test-docr-docker test-docr-multi

# ---------------------------------------------------------------------------
# Documentation (Docusaurus)
# ---------------------------------------------------------------------------

## docs-install: Install documentation dependencies
docs-install:
	cd docs && pnpm install

# Suppress Node.js deprecation warning from Docusaurus/webpack internals (DEP0169).
# This is not in our code; it will be resolved in a future Docusaurus release.
DOCS_NODE_FLAGS := NODE_OPTIONS=--no-deprecation

## docs-dev: Start documentation dev server with hot reload
docs-dev: docs-install
	cd docs && $(DOCS_NODE_FLAGS) pnpm start

## docs-build: Build documentation for production
docs-build: docs-install
	cd docs && $(DOCS_NODE_FLAGS) pnpm run build

## docs-serve: Serve production build locally
docs-serve: docs-build
	cd docs && $(DOCS_NODE_FLAGS) pnpm run serve

# ---------------------------------------------------------------------------
# Release
# ---------------------------------------------------------------------------
# Requires: gh CLI (brew install gh), goreleaser (brew install goreleaser)

RELEASE_VERSION ?= v0.1.0

## release-token: Print gh CLI token and verify access to tap repos
release-token:
	@command -v gh >/dev/null 2>&1 || { echo "gh CLI required: brew install gh"; exit 1; }
	@echo "HOMEBREW_TAP_GITHUB_TOKEN:"
	@echo ""
	@gh auth token
	@echo ""
	@echo "Set this as a repo secret:"
	@echo "  gh secret set HOMEBREW_TAP_GITHUB_TOKEN --repo mistergrinvalds/lazyoci"
	@echo ""
	@echo "Verifying token has access to tap repos..."
	@gh api repos/greenforests-studio/homebrew-tap --jq '.full_name' && echo "  homebrew-tap: OK" || echo "  homebrew-tap: FAILED -- check gh auth scope"
	@gh api repos/greenforests-studio/scoop-bucket --jq '.full_name' && echo "  scoop-bucket: OK" || echo "  scoop-bucket: FAILED -- check gh auth scope"

## release-dry-run: Run goreleaser locally without publishing
release-dry-run: build
	@command -v goreleaser >/dev/null 2>&1 || { echo "goreleaser required: brew install goreleaser"; exit 1; }
	HOMEBREW_TAP_GITHUB_TOKEN=dry-run goreleaser release --snapshot --clean --skip=publish

## release-test: Push all commits, tag RELEASE_VERSION, and trigger the release workflow
release-test:
	@echo "==> Pre-flight checks"
	@command -v gh >/dev/null 2>&1 || { echo "gh CLI required: brew install gh"; exit 1; }
	@git diff --quiet || { echo "error: working tree is dirty -- commit or stash first"; exit 1; }
	@git diff --cached --quiet || { echo "error: staged changes exist -- commit first"; exit 1; }
	@gh secret list --repo mistergrinvalds/lazyoci | grep -q HOMEBREW_TAP_GITHUB_TOKEN || \
		{ echo "error: HOMEBREW_TAP_GITHUB_TOKEN secret not set -- run: make release-token"; exit 1; }
	@echo ""
	@echo "==> Pushing commits to origin..."
	git push origin main
	@echo ""
	@echo "==> Tagging $(RELEASE_VERSION)..."
	git tag -a $(RELEASE_VERSION) -m "Release $(RELEASE_VERSION)"
	git push origin $(RELEASE_VERSION)
	@echo ""
	@echo "==> Release workflow triggered!"
	@echo "    Watch: gh run watch -R mistergrinvalds/lazyoci"
	@echo "    Or:    https://github.com/mistergrinvalds/lazyoci/actions"
	@echo ""
	@echo "    After the run completes, verify:"
	@echo "      gh release view $(RELEASE_VERSION) -R mistergrinvalds/lazyoci"
	@echo "      brew install greenforests-studio/tap/lazyoci"
	@echo "      scoop bucket add greenforests https://github.com/greenforests-studio/scoop-bucket && scoop install lazyoci"
