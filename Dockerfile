FROM golang:1.19-alpine as builder

RUN apk --no-cache add ca-certificates git
WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN go build

RUN chmod +x generate_config.sh
RUN ./generate_config.sh > config.prod.yml

FROM alpine
WORKDIR /
COPY --from=builder /build/job-scheduler .
COPY --from=builder /build/config.prod.yml .
CMD ["./job-scheduler -config config.prod.yml"]
