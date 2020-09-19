FROM golang:1.15.2-buster

ENV GO111MODULE=on
WORKDIR /app
COPY . .
RUN go build
RUN apt-get update
RUN apt-get install -y ca-certificates
COPY webconfig.yaml /webconfig.yaml
COPY config.yaml /config.yaml
COPY data/*.tsv /data/
CMD ["./chinesenotes-go"]