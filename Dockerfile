FROM node:20-alpine AS ui-builder
RUN apk add --no-cache libc6-compat
WORKDIR /app

COPY ui/adminv2 .

RUN npm install

WORKDIR /app/apps/web

RUN npx next build

FROM golang:1.26.2-alpine AS go-builder
RUN apk add --no-cache build-base
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=ui-builder /app/apps/web/out ./pkg/ui/out

ENV CGO_ENABLED=1
RUN go build -o wasmforge cmd/gateway/main.go

FROM alpine:3.22.4 AS runner
WORKDIR /root

COPY --from=go-builder /app/wasmforge ./
EXPOSE 8080
# Defualt port of the proxy server
EXPOSE 9000
ENTRYPOINT ["./wasmforge"]

USER nobody