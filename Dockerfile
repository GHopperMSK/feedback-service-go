from golang:buster as builder

WORKDIR /app
ADD . .
RUN go build -o /usr/local/bin /app/cmd/*

EXPOSE 8080
CMD ["sh", "-c", "/usr/local/bin/kafka-consumer & /usr/local/bin/rest-server"]