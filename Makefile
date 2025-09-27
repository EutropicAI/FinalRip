GO ?= go

.DEFAULT_GOAL := default

IMAGE_NAME := lychee0/finalrip

.PHONY: tidy
tidy:
	${GO} mod tidy

.PHONY: server
server:
	${GO} build -o server ./server

.PHONY: worker
worker:
	${GO} build -o worker ./worker

.PHONY: test
test: tidy
	${GO} test ./...

.PHONY: lint
lint:
	golangci-lint run
	pre-commit install
	pre-commit run --all-files

.PHONY: all
all:
	docker buildx build -f ./deploy/server.dockerfile \
		-t ${IMAGE_NAME}:server \
		-t ${IMAGE_NAME}:server-dev .

	docker buildx build -f ./deploy/worker-cut.dockerfile \
		-t ${IMAGE_NAME}:worker-cut \
		-t ${IMAGE_NAME}:worker-cut-dev .

	docker buildx build -f ./deploy/worker-merge.dockerfile \
		-t ${IMAGE_NAME}:worker-merge \
		-t ${IMAGE_NAME}:worker-merge-dev .

	docker buildx build -f ./deploy/dashboard.dockerfile \
		-t ${IMAGE_NAME}:dashboard \
		-t ${IMAGE_NAME}:dashboard-dev .

.PHONY: pt
pt:
	docker buildx build -f ./deploy/worker-encode.dockerfile --build-arg BASE_CONTAINER_TAG=cuda-dev \
		-t ${IMAGE_NAME}:worker-encode \
		-t ${IMAGE_NAME}:worker-encode-dev \
		-t ${IMAGE_NAME}:worker-encode-cuda-dev \
		-t ${IMAGE_NAME}:worker-encode-cuda .

.PHONY: dev
dev:
	docker compose -f ./deploy/docker-compose/lite/docker-compose.yml down
	docker compose -f ./deploy/docker-compose/lite/docker-compose.yml up -d
