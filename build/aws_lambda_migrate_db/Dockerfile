FROM golang:1.25 AS builder

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -ldflags="-s -w" -o main scripts/aws_lambda_migrate_db/main.go

# AWS certs
RUN wget -O ./global-bundle.pem https://truststore.pki.rds.amazonaws.com/global/global-bundle.pem

FROM public.ecr.aws/lambda/go:latest

COPY --from=builder /app/main /var/task/main
COPY --from=builder /app/configs /var/task/configs
COPY --from=builder /app/internal/database/migrations /var/task/internal/database/migrations
COPY --from=builder /app/global-bundle.pem /var/task/global-bundle.pem

RUN chmod 755 /var/task/main

CMD [ "main" ]
