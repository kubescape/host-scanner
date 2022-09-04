FROM golang:1.18 
WORKDIR /app/
COPY . ./
RUN go get ./...
RUN go test -v ./...
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o kube-host-sensor --ldflags '-w -s'

FROM alpine
COPY --from=0 /app/kube-host-sensor /.
ENTRYPOINT [ "./kube-host-sensor" ]