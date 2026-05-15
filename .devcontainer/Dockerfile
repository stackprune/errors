FROM golang:1.22 AS builder

RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.12.2 && \
    mv $(go env GOPATH)/bin/golangci-lint $(go env GOPATH)/bin/golangci-lint-v2

FROM golang:1.22

ENV LANG=C.UTF-8
ENV LC_CTYPE=C.UTF-8
ENV TZ=Asia/Tokyo

RUN apt-get update && \
  apt-get install -y \
    bash-completion \
    jq \
    less \
    vim && \
  rm -rf /var/lib/apt/lists/*

RUN useradd -m -s /bin/bash nonroot && \
  chown -R nonroot:nonroot /go

USER nonroot
WORKDIR /workspace
COPY --from=builder --chown=nonroot /go/bin /go/bin

CMD ["bash"]
