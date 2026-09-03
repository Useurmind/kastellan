# Build stage with Go and kubebuilder
FROM golang:1.27-bookworm AS builder

# Install system dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    bash \
    curl \
    git \
    make \
    ca-certificates \
    podman \
    unzip \
    && rm -rf /var/lib/apt/lists/*

# Install Go tools
RUN go install github.com/go-delve/delve/cmd/dlv@latest && \
    go install golang.org/x/tools/cmd/goimports@latest && \
    go install golang.org/x/lint/golint@latest && \
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest && \
    go install golang.org/x/tools/cmd/deadcode@latest && \
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Install kubebuilder
RUN curl -L -o kubebuilder https://go.kubebuilder.io/dl/latest/$(go env GOOS)/$(go env GOARCH) && \
    chmod +x kubebuilder && \
    mv kubebuilder /usr/local/bin/

# Install controller-gen
RUN curl -L -o controller-gen https://github.com/kubernetes-sigs/controller-tools/releases/download/v0.21.0/controller-gen-linux-amd64 && \
    chmod +x controller-gen && \
    mv controller-gen /usr/local/bin/

# Install protobuf compiler
RUN curl -L -o protobuf.zip https://github.com/protocolbuffers/protobuf/releases/download/v36.1/protoc-36.1-linux-x86_64.zip && \
    unzip protobuf.zip -d /usr/local && \
    rm protobuf.zip

# Runtime stage
FROM ubuntu:24.04

RUN apt-get update && apt-get install -y \
    git \
    curl \
    make \
    bash \
    podman \
    build-essential \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Copy controller-gen for manifest generation
COPY --from=builder /usr/local/bin/controller-gen /usr/local/bin/controller-gen

# Copy kubebuilder
COPY --from=builder /usr/local/bin/kubebuilder /usr/local/bin/kubebuilder

# Copy go bin tools
COPY --from=builder /go/bin/dlv /usr/local/bin/dlv
COPY --from=builder /go/bin/goimports /usr/local/bin/goimports
COPY --from=builder /go/bin/golint /usr/local/bin/golint
COPY --from=builder /go/bin/golangci-lint /usr/local/bin/golangci-lint
COPY --from=builder /go/bin/deadcode /usr/local/bin/deadcode
COPY --from=builder /go/bin/protoc-gen-go /usr/local/bin/protoc-gen-go
COPY --from=builder /go/bin/protoc-gen-go-grpc /usr/local/bin/protoc-gen-go-grpc

# Copy protobuf compiler
COPY --from=builder /usr/local/bin/protoc /usr/local/bin/protoc
COPY --from=builder /usr/local/include /usr/local/include
COPY --from=builder /usr/local/lib /usr/local/lib

RUN curl -fsSL https://go.dev/dl/go1.27.0.linux-amd64.tar.gz | tar -C /usr/local -xz

ENV PATH="/usr/local/go/bin:${PATH}"

RUN useradd -m -s /bin/bash agent

USER agent

WORKDIR /home/agent

RUN bash -c 'echo "export PATH=$PATH:/home/agent/go/bin" >> /home/agent/.bashrc'


RUN curl -fsSL https://opencode.ai/install | bash

CMD ["opencode"]
