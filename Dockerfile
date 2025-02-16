FROM oven/bun:1.2.2 AS frontend-build
WORKDIR /build
COPY package.json bun.lock ./
RUN bun install --frozen-lockfile --ignore-scripts --no-progress
ARG VITE_MAP_STYLE_KEY
ENV VITE_MAP_STYLE_KEY=$VITE_MAP_STYLE_KEY
COPY . .
RUN bun run build

FROM golang:1.24.0-bookworm AS go-build
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend-build /build/dist /build/dist
ARG GIT_SHA
ENV GIT_SHA=$GIT_SHA
RUN go build -ldflags "-X main.GitCommit=${GIT_SHA}" -o /build/seed_search cli/seed_search.go
RUN go build -ldflags "-X main.GitCommit=${GIT_SHA}" -o /build/fountain_finder main.go

FROM debian:bookworm-slim AS runtime
RUN apt-get update \
    && apt-get install -y ca-certificates curl \
    && apt-get autoremove -y \
    && apt-get clean -y \
    && rm -rf /var/lib/apt/lists/*
COPY --from=go-build /build/seed_search /build/fountain_finder /build/
EXPOSE 8000
CMD ["/build/fountain_finder"]
HEALTHCHECK --interval=5m --timeout=3s \
    CMD curl -f http://localhost:8000/api/health || exit 1
