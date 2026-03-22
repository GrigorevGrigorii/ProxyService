ARG PORT=8080
ARG WORKDIR="/app"
ARG SERVICE="" # proxy-api-server | mock-api-server

FROM golang:1.25 AS builder

ARG WORKDIR
ARG SERVICE

ENV CGO_ENABLED=0
ENV GOOS=linux

WORKDIR ${WORKDIR}

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o ./app cmd/${SERVICE}/main.go

FROM alpine:3.16 AS production

ARG WORKDIR
ARG PORT

ENV PORT=${PORT}

WORKDIR ${WORKDIR}

RUN apk add --no-cache ca-certificates

COPY --from=builder ${WORKDIR}/app ./app
COPY --from=builder ${WORKDIR}/configs ./configs

EXPOSE ${PORT}

CMD ["./app"]
