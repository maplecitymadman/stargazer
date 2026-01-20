FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY bin/stargazer-linux-amd64 /usr/local/bin/stargazer
RUN chmod 755 /usr/local/bin/stargazer
# Copy frontend static files to /app/frontend/out
COPY frontend/out /app/frontend/out
EXPOSE 8000
# Run from /app so relative paths work
WORKDIR /app
CMD ["stargazer"]
