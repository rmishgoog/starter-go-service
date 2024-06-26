# Check to see if we can use ash, in Alpine images, or default to BASH.
SHELL_PATH = /bin/ash
SHELL = $(if $(wildcard $(SHELL_PATH)),/bin/ash,/bin/bash)

# =========================================================================
# Define the needed environment variables

GOLANG          := golang:1.22
ALPINE          := alpine:3.19
KIND			:= kindest/node:v1.29.2
KIND_CLUSTER	:= local-starter-cluster
POSTGRES		:= postgres:16.2
GRAFANA			:= grafana/grafana:10.4.0
PROMETHEUS      := prom/prometheus:v2.51.0
TEMPO           := grafana/tempo:2.4.0
LOKI            := grafana/loki:2.9.0
PROMTAIL        := grafana/promtail:2.9.0
AUTH_APP        := auth
SALES_APP		:= sales
NAMESPACE		:= sales-system
BASE_IMAGE_NAME := gcr.io/rmishra-kubernetes-playground
VERSION			:= 0.0.2
SALES_IMAGE		:= $(BASE_IMAGE_NAME)/$(SALES_APP):$(VERSION)
METRICS_IMAGE   := $(BASE_IMAGE_NAME)/metrics:$(VERSION)
AUTH_IMAGE      := $(BASE_IMAGE_NAME)/$(AUTH_APP):$(VERSION)
# ===========================================================================

# ===========================================================================
# Echo the environment variables

echo-env:
	 @echo $(SALES_APP)
	 @echo $(VERSION)
	 @echo $(SALES_IMAGE)

# ===========================================================================
# Run locally & do log formatting
run:
	go run app/services/sales-api/main.go | go run app/tooling/logfmt/main.go


run-help:
	go run app/services/sales-api/main.go --help | go run app/tooling/logfmt/main.go
# ===========================================================================

# ===========================================================================
# Generate a private key for the purpose of signing the jwt token
run-keygen:
	go run app/tooling/sales-admin/main.go
# ============================================================================
# Use openssl to generate the public-private key pair instead
run-openssl:
	openssl genpkey -algorithm RSA -out private.pem -pkeyopt rsa_keygen_bits:2048
	openssl rsa -pubout -in private.pem -out public.pem

# ===========================================================================
# Start or stop a local kind kubernetes cluster (local environment bootstrap or tear down)
dev-up:
	kind create cluster \
		--image $(KIND) \
		--name $(KIND_CLUSTER) \
		--config zarf/k8s/dev/kind-config.yaml
	
	kubectl wait --timeout=120s --namespace=local-path-storage --for=condition=Available deployment/local-path-provisioner

dev-down:
	kind delete cluster --name $(KIND_CLUSTER)
# ============================================================================

# ============================================================================
# Building container images

build:	sales

sales:
	docker build \
		-f zarf/docker/dockerfile.service \
		-t $(SALES_IMAGE) \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
		.
# =============================================================================

# =============================================================================
# Set up the application on development (local kind) cluster

bootsrap-app:	dev-load dev-apply dev-status dev-logs

dev-load:
	kind load docker-image $(SALES_IMAGE) --name $(KIND_CLUSTER)

dev-apply:
	kustomize build zarf/k8s/dev/sales | kubectl apply -f -

dev-status:
	kubectl wait --for=condition=Ready --timeout=30s --namespace=$(NAMESPACE) pods -l app=$(SALES_APP)

dev-logs:
	kubectl logs --namespace=$(NAMESPACE) -l app=$(SALES_APP) --all-containers=true -f --tail=100 --max-log-requests=6 | go run app/tooling/logfmt/main.go

dev-restart:
	kubectl rollout restart deployment $(SALES_APP) --namespace=$(NAMESPACE)
# ==============================================================================

# ==============================================================================
# Go mod management

tidy:
	go mod tidy
	go mod vendor

# ===============================================================================
# Local curl on localhost

curl:
	curl -il http://localhost:3000/hack
# ===============================================================================

# ===============================================================================
# Build image, upload and restart the deployment. This is useful for quick build &
# image uploads, followed by container restart

dev-update:	build	dev-load	dev-apply	dev-restart
# ===============================================================================