FROM golang:1.17 as builder
WORKDIR /app
 
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download                                                    
 
RUN apt -y update && apt install --yes --no-cache ca-certificates git
RUN update-ca-certificates
ADD . /app/
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags="-s -w" -o app .
    

FROM gcr.io/distroless/static:nonroot
WORKDIR /
 
COPY --from=builder /app/backend .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
USER 65532:65532
 
ENTRYPOINT ["./app"]

