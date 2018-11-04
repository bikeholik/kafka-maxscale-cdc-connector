FROM golang:1.11.1 AS build
COPY . /go/src/github.com/bborbe/kafka-maxscale-cdc-connector
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s" -a -installsuffix cgo -o /main ./src/github.com/bborbe/kafka-maxscale-cdc-connector
CMD ["/bin/bash"]

FROM scratch
COPY --from=build /main /main
ADD https://curl.haxx.se/ca/cacert.pem /etc/ssl/certs/ca-certificates.crt
ENTRYPOINT ["/main"]
