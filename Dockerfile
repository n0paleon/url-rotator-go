# stage 1: build the binary
FROM golang:1.22.5-alpine AS builder
WORKDIR /app
# install build dependencies
RUN apk add --no-cache git
# opy all source code and project files to working directory
COPY . .
# download required go modules
RUN go mod download
# build the binary
RUN mkdir -p bin
RUN go build -ldflags="-s -w" -o bin/app cmd/web/main.go
# build migrator binary
RUN go build -ldflags="-s -w" -o bin/migrate cmd/migrate/main.go

# stage 2: runtime image
FROM alpine:latest
WORKDIR /app
# install runtime dependencies
RUN apk add --no-cache dumb-init
# copy the binary and project files from the build stage
COPY --from=builder /app .
# Make the entrypoint script executable
RUN chmod +x /app/entrypoint.sh
# create default migration status
RUN echo "0" >> /app/migrated
# set the entrypoint and command
ENTRYPOINT ["/usr/bin/dumb-init", "--", "/app/entrypoint.sh"]
