FROM golang:1.21-alpine as builder

WORKDIR /build

# Copy depends, those don't change so often.
COPY go.mod .
COPY go.sum .

RUN go mod download

# Copy code to build the application.
COPY . .
RUN go build -o delay-kafka main.go

FROM alpine

WORKDIR /app

RUN apk update && \
    apk add --no-cache tzdata

COPY --from=builder /usr/local/go/lib/time/zoneinfo.zip /usr/local/go/lib/time/zoneinfo.zip
COPY --from=builder /build/delay-kafka .

ENTRYPOINT ["/app/delay-kafka"]