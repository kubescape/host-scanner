FROM golang:1.17
WORKDIR /app/
COPY . ./
RUN go get ./...
RUN go build -o kube-host-sensor



FROM scratch
COPY --from=0 /app/kube-host-sensor /.
ENTRYPOINT [ "kube-host-sensor" ]