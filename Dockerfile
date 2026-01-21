# Use official Python runtime as a parent image
FROM python:3.11-slim

# Set working directory
WORKDIR /app

# Install system dependencies
# We might need gcc or other build tools if some python packages require them,
# but for now we'll stick to basic slim image.
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements
COPY requirements.txt .

# Install dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Copy source code
COPY src/ src/

# Expose port (if web server is used, default 8000)
EXPOSE 8000

# Run the application
# We use -m src.main to support relative imports within the package
ENTRYPOINT ["python", "-m", "src.main"]
