# Two builders, one runtime image. The frontend is compiled into the Go
# binary, so what ships is a single static file plus a CA bundle.

FROM node:20-alpine AS frontend
WORKDIR /build
# Copy the manifests first so `npm ci` is cached until a dependency
# actually changes.
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
# angular.json already points the build at ../internal/webui/dist, which
# from /build resolves to /internal/webui/dist. Leaving that alone keeps
# the container build and a local `make build` producing the same layout.
RUN npm run build

FROM golang:1.26-alpine AS backend
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /internal/webui/dist/ ./internal/webui/dist/
ARG VERSION=dev
RUN CGO_ENABLED=0 go build \
    -trimpath \
    -ldflags "-s -w -X github.com/statusengine/interface/internal/httpapi.Version=${VERSION}" \
    -o /out/seid ./cmd/seid

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -u 10001 seid
COPY --from=backend /out/seid /usr/local/bin/seid
USER 10001
# Inside a container the loopback default would make the service
# unreachable from anywhere, so the image binds all interfaces and leaves
# exposure to the port mapping.
ENV SEI_LISTEN_ADDR=:8090
EXPOSE 8090
ENTRYPOINT ["/usr/local/bin/seid"]
CMD ["serve"]
