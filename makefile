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
VERSION			:= 0.0.1
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
# ===========================================================================

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

bootsrap-app:	dev-load dev-apply dev-logs

dev-load:
	kind load docker-image $(SALES_IMAGE) --name $(KIND_CLUSTER)

dev-apply:
	kustomize build zarf/k8s/dev/sales | kubectl apply -f -
#	kubectl wait pods --namespace=$(NAMESPACE) --selector app=$(SALES_APP) --timeout=120s --for=condition=Ready

dev-status:
	watch -n 2 kubectl get pods -o wide --all-namespaces

dev-logs:
	kubectl logs --namespace=$(NAMESPACE) -l app=$(SALES_APP) --all-containers=true -f --tail=100 --max-log-requests=6 | go run app/tooling/logfmt/main.go

dev-restart:
	kubectl rollout restart deployment $(SALES_APP) --namespace=$(NAMESPACE)
# ==============================================================================
