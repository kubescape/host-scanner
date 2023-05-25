FROM golang:1.19 
ARG BUILD_VERSION
WORKDIR /app/
COPY . ./
ENV CGO_ENABLED=0
RUN go get ./...
RUN go build -o kube-host-sensor --ldflags "-w -s -X main.BuildVersion=$BUILD_VERSION"

FROM alpine
COPY --from=0 /app/kube-host-sensor /.
ENTRYPOINT [ "./kube-host-sensor" ]
