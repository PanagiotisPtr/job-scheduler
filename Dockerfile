FROM golang:1.19-alpine

RUN apk --no-cache add ca-certificates git
WORKDIR $GOPATH/github.com/panagiotisptr

RUN git clone https://github.com/PanagiotisPtr/job-scheduler.git

WORKDIR $GOPATH/github.com/panagiotisptr/job-scheduler

RUN go mod download

RUN cd cmd/job-scheduler && go install

CMD ["job-scheduler"]
