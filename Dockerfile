FROM golang:1.18

RUN apt update && apt install libssl-dev

WORKDIR /project
COPY . .

RUN go mod download && go mod vendor
# workaround for missing SDK files
RUN cp -R tmp/c vendor/github.com/hyperledger/sawtooth-sdk-go/c

RUN go build -o app cmd/main.go
RUN chmod +x app
ENTRYPOINT ["/project/app"]
