# Quantify

Quantify is a comprehensive inventory management system designed to streamline and automate your inventory processes. It provides a robust set of features to manage products, track stock levels, handle orders, and gain insights into your business performance.

## Core Features

- **Product Management:** Easily create, update, and manage your product catalog, including categories, sub-categories, and suppliers.
- **Stock Control:** Keep your inventory accurate with features for stock adjustments, history tracking, and seamless stock transfers between different locations.
- **Order Management:** Efficiently manage your procurement process with tools for creating, approving, sending, and receiving purchase orders.
- **User Management:** Securely manage user access with features for user registration, login, and role-based permissions.
- **Insightful Reporting:** Make data-driven decisions with a variety of reports, including sales trends, inventory turnover, profit margins, and supplier performance.
- **Automated Alerts:** Stay informed about critical inventory events with automated alerts for low stock, expiring products, and other custom triggers.
- **Bulk Operations:** Save time and reduce manual effort with bulk import and export capabilities for your product data.
- **Barcode Scanning:** Streamline your workflow with barcode support for quick product lookups and inventory updates.
- **Demand Forecasting:** Optimize your stock levels and avoid stockouts with intelligent demand forecasting and reorder suggestions.
- **Real-time Notifications:** Stay up-to-date with real-time notifications on important events, such as order status changes and new alerts.

## Technical Overview

The backend of Quantify is built with Go and utilizes the following technologies:

- **Framework:** Gin Gonic
- **Database:** PostgreSQL
- **Cache & Job Queue:** Redis
- **Message Broker:** RabbitMQ for asynchronous task processing
- **Authentication:** JWT-based authentication
- **API:** A versioned RESTful API (`/api/v1`)

## API Documentation

The API is documented using the OpenAPI specification. The `openapi.yaml` file can be found in the `backend/api` directory.

## Getting Started

To get started with the Quantify backend, you will need to have Go, Docker, and Docker Compose installed.

1. **Clone the repository:**
   ```bash
   git clone https://github.com/your-username/quantify.git
   ```
2. **Set up the environment:**
   - Navigate to the `backend` directory.
   - Copy the `.env.example` file to a new file named `.env`.
   - Update the `.env` file with your local configuration for the database, Redis, MinIO, cache TTLs, and other services.
     - **MinIO:** configure `MINIO_ENDPOINT`, credentials, bucket name, and `MINIO_USE_TLS` (set to `true` when pointing at an HTTPS endpoint).
     - **Reporting cache TTLs:** adjust `SALES_TRENDS_CACHE_TTL`, `INVENTORY_TURNOVER_CACHE_TTL`, and `PROFIT_MARGIN_CACHE_TTL` using Go duration strings (`30m`, `2h`, etc.) to control how long each report response is cached.
3. **Run the application:**
   - Use Docker Compose to build and run the services:
     ```bash
     docker-compose up --build
     ```
   - The backend server will be available at `http://localhost:8080`.

## Running the Backend

Once the application is running, you can interact with the API using a tool like Postman or by running the frontend application. The API endpoints are available under the `/api/v1` path.

## CSRF Protection

All authenticated requests that modify state (`POST`, `PUT`, `PATCH`, `DELETE`) require an `X-CSRF-Token` header. The token is returned alongside the access/refresh tokens during login and refresh workflows. Clients must store it securely and include it on every subsequent unsafe request; tokens can be rotated by calling the refresh endpoint and are invalidated on logout.
