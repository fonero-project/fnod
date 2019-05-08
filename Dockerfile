FROM golang:1.11

WORKDIR /go/src/github.com/fonero-project/fnod
COPY . .

RUN env GO111MODULE=on go install . ./cmd/...

EXPOSE 9208

CMD [ "fnod" ]
