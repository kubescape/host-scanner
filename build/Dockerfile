FROM --platform=$BUILDPLATFORM golang:1.20-bullseye as builder
ARG BUILD_VERSION
ENV GO111MODULE=on CGO_ENABLED=0
WORKDIR /work
ARG TARGETOS TARGETARCH

RUN --mount=target=. \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /out/kube-host-sensor --ldflags "-w -s -X main.BuildVersion=$BUILD_VERSION" .

FROM gcr.io/distroless/static-debian11:latest

COPY --from=builder /out/kube-host-sensor /usr/bin/kube-host-sensor

ARG image_version
ENV RELEASE=$image_version

WORKDIR /root
ENTRYPOINT ["kube-host-sensor"]
