FROM golang:1.15

WORKDIR /go/src/github.com/jollyjerr/cryptofu
COPY . .

RUN go build -o cryptofu .

CMD ./cryptofu
