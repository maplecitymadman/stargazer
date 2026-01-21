# Stage 1: Build Backend
FROM golang:alpine AS backend-builder
WORKDIR /app
COPY . .
ENV GOTOOLCHAIN=auto
RUN go build -ldflags "-s -w" -o bin/stargazer cmd/stargazer/main.go

# Stage 2: Build Frontend
FROM node:18-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ .
RUN npm run build

# Stage 3: Final Image
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app

# Copy binary from backend builder
COPY --from=backend-builder /app/bin/stargazer /usr/local/bin/stargazer
RUN chmod 755 /usr/local/bin/stargazer

# Copy frontend assets from frontend builder
COPY --from=frontend-builder /app/frontend/out /app/frontend/out

EXPOSE 8000
CMD ["stargazer"]
