FROM golang:1.25.3 AS builder

WORKDIR /app

COPY go.mod go.sum main.go .

RUN go get
RUN sed -i 's/127.4.4.4/0.0.0.0/g' main.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/main

FROM scratch

COPY --from=builder /app/main /main

EXPOSE 53/udp

ENTRYPOINT ["/main"]
