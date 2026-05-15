REG            ?= docker.io
ORG            ?= jinxf0120
NAME           ?= rancher-exporter
VERSION        ?= v1.0.0
IMAGE          ?= $(REG)/$(ORG)/$(NAME):$(VERSION)

HTTP_PROXY     ?=
HTTPS_PROXY    ?=
NO_PROXY       ?=
GOPROXY        ?=

.PHONY: build
build:
	docker build \
		$(if $(HTTP_PROXY),--build-arg http_proxy=$(HTTP_PROXY),) \
		$(if $(HTTPS_PROXY),--build-arg https_proxy=$(HTTPS_PROXY),) \
		$(if $(NO_PROXY),--build-arg no_proxy=$(NO_PROXY),) \
		$(if $(GOPROXY),--build-arg GOPROXY=$(GOPROXY),) \
		-t $(IMAGE) .

.PHONY: push
push:
	docker push $(IMAGE)

.PHONY: run
run:
	docker run --rm -p 8080:8080 $(IMAGE)

.PHONY: helm-package
helm-package:
	helm package chart/rancher-exporter -d ./dist/

.PHONY: helm-install
helm-install:
	helm upgrade --install rancher-exporter chart/rancher-exporter \
		--namespace cattle-system-exporter --create-namespace

.PHONY: helm-uninstall
helm-uninstall:
	helm uninstall rancher-exporter --namespace cattle-system-exporter
