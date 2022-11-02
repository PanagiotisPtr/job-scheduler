FROM golang:1.19-alpine

RUN apk --no-cache add ca-certificates git
WORKDIR $GOPATH/github.com/panagiotisptr/job-scheduler

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN go build -o job-scheduler cmd/job-scheduler/main.go

RUN chmod +x init.sh
RUN chmod +x generate_config.sh

CMD ["./job-scheduler"]
