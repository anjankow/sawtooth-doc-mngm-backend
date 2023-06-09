FROM golang:1.18

RUN apt update && apt install -y libssl-dev libzmq3-dev

WORKDIR /project
COPY . .

RUN go mod download && go mod vendor
# workaround for missing SDK files
RUN cp -R tmp/c vendor/github.com/hyperledger/sawtooth-sdk-go/c

RUN go build -o app cmd/main.go && chmod +x app

ENTRYPOINT ["/project/app"]
