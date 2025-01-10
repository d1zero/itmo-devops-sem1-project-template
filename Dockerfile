FROM golang:1.23.3-alpine3.19 AS builder

ENV CGO_ENABLED 0

ENV GOOS linux

RUN apk update --no-cache && apk add --no-cache tzdata git curl

WORKDIR /build

ADD go.mod .

ADD go.sum .

RUN go mod download

COPY . .

RUN go build -ldflags="-s -w" -o /build/main cmd/api/main.go
RUN ls

RUN git clone https://github.com/pressly/goose.git
RUN cd goose && go mod tidy && go build -o /build/goose_bin ./cmd/goose
FROM alpine

RUN apk update --no-cache && apk add --no-cache ca-certificates git

WORKDIR /app

COPY --from=builder /build/main /app/main
COPY --from=builder /build/migrate.sh /app/migrate.sh
COPY --from=builder /build/db /app/db
COPY --from=builder /build/goose_bin /app/goose

RUN chmod +x /app/migrate.sh
ENTRYPOINT ["sh","/app/migrate.sh"]
CMD ["./main"]



