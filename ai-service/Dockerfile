FROM python:3.12-slim

WORKDIR /app

# Install poetry
RUN pip install poetry

# Copy configuration files
COPY pyproject.toml poetry.lock ./

# Install dependencies
# --no-root is important because we don't have a package structure
RUN poetry config virtualenvs.create false \
    && poetry install --no-interaction --no-ansi --no-root

# Copy application code
COPY . .

# Expose port
EXPOSE 8000

# Run the application
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
