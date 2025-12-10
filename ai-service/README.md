# Quantify AI Service

This service provides AI-powered insights and forecasting analysis for the Quantify platform, utilizing the Z.AI API.

## Prerequisites

- Python 3.12+
- Poetry

## Installation

1.  Install dependencies:
    ```bash
    poetry install
    ```

2.  Configure environment:
    Copy `.env.example` to `.env` (or create `.env`) and set your API keys:
    ```
    ZAI_API_KEY=your_key
    ZAI_BASE_URL=https://api.z.ai/api/paas/v4/
    ```

## Running the Service

```bash
poetry run python main.py
```
Or directly with uvicorn:
```bash
poetry run uvicorn main:app --reload --port 8001
```

## Endpoints

- `POST /insight`: Generate business insights.
- `POST /analyze-forecast`: Analyze inventory forecast data.
