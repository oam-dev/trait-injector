
# Image URL to use all building/pushing image targets
IMG ?= oam-dev/trait-injector:v1

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: build

# Run tests
test: generate fmt vet manifests kubebuilder
	go test `go list ./... | grep -v e2e-test` -coverprofile cover.out

build:
	mkdir -p bin/
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go

# Install CRDs into a cluster
install: manifests
	kustomize build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests
	kustomize build config/crd | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default | kubectl apply -f -

# Don't forget to `eval $(minikube docker-env)`
minikube: generate manifests docker-build
	kubectl apply -f config/crd/bases/core.oam.dev_servicebindings.yaml
	kubectl delete -f example/manager.yaml || true
	kubectl apply -f example/manager.yaml

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths="./..."

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Build the docker image
# first time needs to do `make` in ssl/
docker-build:
	eval $(minikube docker-env)
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.4 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

# kubebuilder binary
kubebuilder:
ifeq (, $(shell which kubebuilder))
	# Download kubebuilder and extract it to tmp
	curl -sL https://go.kubebuilder.io/dl/2.3.0/$(shell go env GOOS)/$(shell go env GOARCH) | tar -xz -C /tmp/
	# You'll need to set the KUBEBUILDER_ASSETS env var if you put it somewhere else
	sudo mv /tmp/kubebuilder_2.3.0_$(shell go env GOOS)_$(shell go env GOARCH) /usr/local/kubebuilder
newPATH:=$(PATH):/usr/local/kubebuilder/bin
export PATH=$(newPATH)
endif

kind-e2e:
	docker build -t $(IMG) -f Dockerfile .
	kind load docker-image $(IMG) \
		|| { echo >&2 "kind not installed or error loading image: $(IMG)"; exit 1; } && \
	kubectl apply -f ./config/crd/bases/core.oam.dev_servicebindings.yaml
	./charts/injector/gen_certs.sh e2e-trait-injector
	helm version
	helm install e2e ./charts/injector --set image.repository=$(IMG) --wait \
		|| { echo >&2 "helm install timeout"; kubectl logs `kubectl get pods -l "app.kubernetes.io/name=rudr,app.kubernetes.io/instance=rudr" -o jsonpath="{.items[0].metadata.name}"`; exit 1; } && \
	go test -v ./e2e-test/
