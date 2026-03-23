FROM golang:1.25 AS builder

ARG SERVICE # proxy-api-server | mock-api-server | admin-api-server

ENV CGO_ENABLED=0
ENV GOOS=linux

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o ./bin/main cmd/${SERVICE}/main.go

FROM alpine:3.23 AS service

ARG PORT

ENV PORT=${PORT}

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/bin/main ./main
COPY --from=builder /app/configs ./configs

EXPOSE ${PORT}

CMD ["./main"]
