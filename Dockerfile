FROM golang:1.26.4 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -o /app/server \
  cmd/server/main.go


FROM gcr.io/distroless/static-debian13:nonroot
WORKDIR /app
COPY --from=builder /app/server server
ENV SUBSERV_HTTP_HOST=localhost
ENTRYPOINT ["/app/server"]
