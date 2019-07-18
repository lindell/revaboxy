FROM golang:latest as builder
WORKDIR /go/src/github.com/lindell/revaboxy
COPY cmd cmd
COPY pkg pkg
COPY internal internal
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o app cmd/revaboxy/main.go

FROM scratch
COPY --from=builder /go/src/github.com/lindell/revaboxy/app .
ENTRYPOINT ["./app"]
