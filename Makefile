VERSION ?= $(shell git tag -l --sort=v:refname | tail -1)
GIT_COMMIT := $(shell git describe --match=NeVeRmAtCh --always --abbrev=40)
BUILD_TIME := $(shell date +"%Y-%m-%dT%H:%M:%SZ")
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)

GOOS := $(shell go env GOHOSTOS)
GOARCH := $(shell go env GOHOSTARCH)
TARGET := ws-${GOOS}-${GOARCH}
OS_ARCH := ${GOOS}/${GOARCH}

SERVER_DIR := ./server
CLIENT_DIR := ./client
SERVER_NAME := ws-server
CLIENT_NAME := ws-client
OUTPUT_DIR := ./bin

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS=-ldflags="-X 'ws-example/common.Version=${GOOS}-${GOARCH}-${GIT_COMMIT}'"

GO111MODULE=on
GOPROXY=https://goproxy.cn,direct

.PHONY: all
all: ws-all

.PHONY: ws-all
ws-all: darwin-amd64 darwin-arm64 \
linux-amd64 linux-arm64

# ---------darwin-----------
.PHONY: darwin-amd64
darwin-amd64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o $(OUTPUT_DIR)/darwin-amd64-$(SERVER_NAME) ${SERVER_DIR}
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o $(OUTPUT_DIR)/darwin-amd64-$(CLIENT_NAME) ${CLIENT_DIR}
	chmod +x $(OUTPUT_DIR)/darwin-amd64-$(SERVER_NAME)
	chmod +x $(OUTPUT_DIR)/darwin-amd64-$(CLIENT_NAME)
.PHONY: darwin-arm64
darwin-arm64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build ${LDFLAGS} -o $(OUTPUT_DIR)/darwin-arm64-$(SERVER_NAME) ${SERVER_DIR}
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build ${LDFLAGS} -o $(OUTPUT_DIR)/darwin-arm64-$(CLIENT_NAME) ${CLIENT_DIR}
	chmod +x $(OUTPUT_DIR)/darwin-arm64-$(SERVER_NAME)
	chmod +x $(OUTPUT_DIR)/darwin-arm64-$(CLIENT_NAME)
# ---------darwin-----------

# ---------linux-----------
.PHONY: linux-amd64
linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o $(OUTPUT_DIR)/linux-amd64-$(SERVER_NAME) ${SERVER_DIR}
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o $(OUTPUT_DIR)/linux-amd64-$(CLIENT_NAME) ${CLIENT_DIR}
	chmod +x $(OUTPUT_DIR)/linux-amd64-$(SERVER_NAME)
	chmod +x $(OUTPUT_DIR)/linux-amd64-$(CLIENT_NAME)
.PHONY: linux-arm64
linux-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build ${LDFLAGS} -o $(OUTPUT_DIR)/linux-arm64-$(SERVER_NAME) ${SERVER_DIR}
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build ${LDFLAGS} -o $(OUTPUT_DIR)/linux-arm64-$(CLIENT_NAME) ${CLIENT_DIR}
	chmod +x $(OUTPUT_DIR)/linux-arm64-$(SERVER_NAME)
	chmod +x $(OUTPUT_DIR)/linux-arm64-$(CLIENT_NAME)
# ---------linux-----------