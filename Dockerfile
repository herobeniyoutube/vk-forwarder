FROM golang:1.25.6 AS build
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
COPY . ./
RUN mkdir -p /app/tmp
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates curl zlib1g \
    && rm -rf /var/lib/apt/lists/*
ARG TARGETARCH
RUN if [ "$TARGETARCH" = "arm64" ]; then \
      YT_BIN="yt-dlp_linux_aarch64"; \
    else \
      YT_BIN="yt-dlp_linux"; \
    fi \
    && curl -L "https://github.com/yt-dlp/yt-dlp/releases/latest/download/${YT_BIN}" -o /usr/local/bin/yt-dlp \
    && chmod +x /usr/local/bin/yt-dlp
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o vk-forwarder ./

FROM debian:trixie-slim
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates ffmpeg zlib1g \
    && rm -rf /var/lib/apt/lists/*
RUN useradd -r -u 65532 -g nogroup nonroot
WORKDIR /app
ENV PATH="/usr/local/bin:${PATH}"
COPY --from=build --chown=nonroot:nogroup /app/vk-forwarder /app/vk-forwarder
COPY --from=build --chown=nonroot:nogroup /app/tmp /app/tmp
COPY --from=build /usr/local/bin/yt-dlp /usr/local/bin/yt-dlp
USER nonroot
EXPOSE 14888
ENTRYPOINT ["/app/vk-forwarder"]
