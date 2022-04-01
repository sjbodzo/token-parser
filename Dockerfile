FROM golang:1.17 as builder
WORKDIR /app

ENV DEBIAN_FRONTEND=noninteractive
 
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download                                                    

RUN apt update && apt install --yes ca-certificates
RUN update-ca-certificates
ADD . /app
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags="-s -w" -o coinparser .
    

FROM gcr.io/distroless/static:nonroot
WORKDIR /

COPY --from=builder /app/ .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
USER 65532:65532
 
ENTRYPOINT ["./coinparser", "server"]

