FROM --platform=$BUILDPLATFORM golang:1.23.10-bookworm AS builder

ENV GO111MODULE=on CGO_ENABLED=0
WORKDIR /work
ARG BUILD_VERSION
ARG TARGETOS TARGETARCH

RUN --mount=target=. \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /out/kube-host-sensor --ldflags "-w -s -X main.BuildVersion=$BUILD_VERSION" .

FROM gcr.io/distroless/static-debian12:latest

COPY --from=builder /out/kube-host-sensor /usr/bin/kube-host-sensor

ARG image_version
ENV RELEASE=$image_version

WORKDIR /root
ENTRYPOINT ["kube-host-sensor"]
