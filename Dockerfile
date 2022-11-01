FROM golang:1.19-alpine as builder

RUN apk --no-cache add ca-certificates git
WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN go build -o job-scheduler cmd/job-scheduler/main.go
RUN chmod +x generate_config.sh

FROM alpine
WORKDIR /
COPY --from=builder /build/job-scheduler .
COPY --from=builder /build/generate_config.sh .
CMD ["./generate_config.sh > config.prod.yml && ./job-scheduler -config config.prod.yml"]
