# Multi-stage build for minimal image
FROM python:3.11-slim as builder

# Install build dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    && rm -rf /var/lib/apt/lists/*

# Install Python dependencies
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir --user -r requirements.txt

# Production stage
FROM python:3.11-slim

# Create non-root user
RUN addgroup --system stargazer && adduser --system --group stargazer

# Install runtime dependencies only
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Copy installed packages from builder
COPY --from=builder /root/.local /home/stargazer/.local

# Create application directory
WORKDIR /app

# Copy application code
COPY src/ ./src/

# Set permissions
RUN chown -R stargazer:stargazer /app

# Switch to non-root user
USER stargazer

# Add .local to PATH
ENV PATH=/home/stargazer/.local/bin:$PATH
ENV PYTHONPATH=/app/src:$PYTHONPATH

# Expose data volume
VOLUME ["/data"]

# Default command
CMD ["python", "src/main.py", "--help"]