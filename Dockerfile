# Use minimal base image
FROM alpine:3.19

# Install ca-certificates for HTTPS requests, shadow for usermod/groupmod, and su-exec
RUN apk --no-cache add ca-certificates shadow su-exec

WORKDIR /app

# Copy the pre-built binary
COPY releases/pman-server-linux-amd64 pman-server

# Create initial user and group with default IDs
RUN addgroup -g 987 pman && \
    adduser -D -u 987 -G pman -s /bin/sh pman

# Create data directory
RUN mkdir -p /data

# Copy entrypoint script
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

# Set environment variable for database path
ENV PMAN_DB_PATH=/data/pman.db

# Default UID/GID (can be overridden at runtime)
ENV PMAN_UID=987
ENV PMAN_GID=987

# Expose port
EXPOSE 8080

# Use entrypoint script
ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["./pman-server"]