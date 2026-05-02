FROM ubuntu:latest
LABEL authors="mikhai5545"

ENTRYPOINT ["top", "-b"]
FROM node:18-alpine AS ui-builder
WORKDIR /app
COPY ui/adminv2/package*.json ./
COPY ui/adminv2/ ./
RUN npm install
RUN npm run build

FROM golang:1.25-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=ui-builder /app/apps/web/out ./pkg/ui/out
RUN go build -o gateway cmd/gateway/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=go-builder /app/gateway ./
EXPOSE 8080
EXPOSE 9090
CMD ["./gateway"]
