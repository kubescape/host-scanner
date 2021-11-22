FROM golang:1.17
WORKDIR /app/
COPY . ./
RUN go get ./...
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o kube-host-sensor --ldflags '-w -s'



FROM scratch
COPY --from=0 /app/kube-host-sensor /.
# ENTRYPOINT [ "sh" ]
ENTRYPOINT [ "./kube-host-sensor" ]