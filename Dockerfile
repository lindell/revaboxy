FROM golang:latest as builder
WORKDIR /go/src/github.com/lindell/revaboxy
COPY cmd cmd
COPY pkg pkg
RUN CGO_ENABLED=0 GOOS=linux go build -o app cmd/revaboxy/main.go

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /
COPY --from=builder /go/src/github.com/lindell/revaboxy/app .
CMD ["./app"]
