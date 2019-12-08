FROM golang:1.13-buster

ENV GO111MODULE=on
WORKDIR /app
COPY . .
RUN go build
RUN apt-get update
RUN apt-get install -y ca-certificates
RUN mkdir -p data
CMD ["./chinesenotes-go"]