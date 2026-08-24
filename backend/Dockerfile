# syntax=docker/dockerfile:1

# ---- build stage ----
FROM golang:1.26 AS build
WORKDIR /src

# Cache dependency download
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# CGO tidak diperlukan (modernc.org/sqlite = pure Go)
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/server ./cmd/server

# ---- runtime stage ----
FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=build /out/server /app/server

ENV PORT=8080 \
    DATABASE_PATH=/app/data/printku.db \
    UPLOAD_DIR=/app/uploads \
    APP_ENV=production

EXPOSE 8080
VOLUME ["/app/data", "/app/uploads"]
ENTRYPOINT ["/app/server"]
