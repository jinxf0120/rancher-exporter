FROM golang:1.19-alpine AS builder

ARG http_proxy
ARG https_proxy
ARG no_proxy
ENV http_proxy=${http_proxy} https_proxy=${https_proxy} no_proxy=${no_proxy}

ARG GOPROXY=
ENV GOPROXY=${GOPROXY}

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o prometheus-rancher-exporter .

FROM alpine:3

RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/prometheus-rancher-exporter .
EXPOSE 8080
CMD ["./prometheus-rancher-exporter"]
