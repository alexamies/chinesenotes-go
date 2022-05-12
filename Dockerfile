FROM golang:1.18.2

ENV GO111MODULE=on
WORKDIR /app
COPY . .
RUN go build
RUN apt-get update
RUN apt-get install -y ca-certificates
COPY webconfig.yaml /webconfig.yaml
COPY config.yaml /config.yaml
COPY data/*.txt /data/
COPY web-resources/*.html /web-resources/
COPY web/* /web/
CMD ["./chinesenotes-go"]