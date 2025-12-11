import os
from dotenv import load_dotenv

load_dotenv()

import httpx
from typing import Optional, Dict, Any

class BackendClient:
    def __init__(self, base_url: Optional[str] = None):
        self.base_url = base_url or os.getenv("BACKEND_URL", "http://localhost:8080")
        self.username = os.getenv("AI_USER_USERNAME", "ai-agent")
        self.email = os.getenv("AI_USER_EMAIL", "ai-agent@quantify.com")
        self.password = os.getenv("AI_USER_PASSWORD", "rZQ$4Rs!6{QHaR{5Sra{]z_%n")
        self.token = None
        self.csrf_token = None
        self.client = httpx.AsyncClient(timeout=30.0)

    async def _login(self):
        """Authenticate and retrieve a new token."""
        url = f"{self.base_url}/api/v1/users/login"
        try:
            response = await self.client.post(url, json={"username": self.username, "password": self.password})
            response.raise_for_status()
            data = response.json()
            self.token = data.get("accessToken")
            self.csrf_token = data.get("csrfToken")
            print(f"AI Agent authenticated as {self.email}")
        except Exception as e:
            print(f"Failed to authenticate AI Agent: {e}")
            # Fallback to manual token if login fails (e.g. dev mode)
            self.token = os.getenv("BACKEND_API_TOKEN")
            self.csrf_token = None

    def _get_headers(self) -> Dict[str, str]:
        headers = {"Content-Type": "application/json"}
        if self.token:
            headers["Authorization"] = f"Bearer {self.token}"
        if self.csrf_token:
            headers["X-CSRF-Token"] = self.csrf_token
        return headers

    async def ensure_auth(self):
        if not self.token:
            await self._login()

    async def get_inventory_status(self, product_id: Optional[int] = None) -> Dict[str, Any]:
        """Fetch inventory status for a product or all products."""
        await self.ensure_auth()
        url = f"{self.base_url}/api/v1/products"
        if product_id:
            url = f"{self.base_url}/api/v1/products/{product_id}"
        
        response = await self.client.get(url, headers=self._get_headers())
        response.raise_for_status()
        return response.json()

    async def create_purchase_order(self, product_id: int, quantity: int, supplier_id: int) -> Dict[str, Any]:
        """Create a draft purchase order."""
        await self.ensure_auth()
        url = f"{self.base_url}/api/v1/replenishment/purchase-orders"
        
        # We need to fetch product price first to populate unit price
        product_url = f"{self.base_url}/api/v1/products/{product_id}"
        prod_resp = await self.client.get(product_url, headers=self._get_headers())
        prod_resp.raise_for_status()
        product = prod_resp.json()
        unit_price = product.get("purchasePrice", 0)

        import datetime
        payload = {
            "supplierId": supplier_id,
            "orderDate": datetime.datetime.now().isoformat(),
            "items": [
                {
                    "productId": product_id,
                    "orderedQuantity": quantity,
                    "unitPrice": unit_price
                }
            ]
        }
        
        response = await self.client.post(url, json=payload, headers=self._get_headers())
        response.raise_for_status()
        return response.json()

    async def get_sales_report(self, start_date: str, end_date: str, product_id: Optional[int] = None) -> Dict[str, Any]:
        """Get sales report for a specific date range."""
        await self.ensure_auth()
        url = f"{self.base_url}/api/v1/reports/sales-trends"
        payload = {
            "startDate": f"{start_date}T00:00:00Z",
            "endDate": f"{end_date}T23:59:59Z",
            "groupBy": "daily"
        }
        if product_id:
            payload["productId"] = product_id
            
        response = await self.client.post(url, json=payload, headers=self._get_headers())
        response.raise_for_status()
        return response.json()

    async def trigger_alert_check(self) -> Dict[str, Any]:
        """Trigger the backend to check for alerts."""
        await self.ensure_auth()
        url = f"{self.base_url}/api/v1/alerts/check"
        response = await self.client.post(url, headers=self._get_headers())
        response.raise_for_status()
        return response.json()

    async def get_active_alerts(self) -> Dict[str, Any]:
        """Get all active alerts."""
        await self.ensure_auth()
        url = f"{self.base_url}/api/v1/alerts?status=ACTIVE"
        response = await self.client.get(url, headers=self._get_headers())
        response.raise_for_status()
        return response.json()

    async def broadcast_notification(self, title: str, message: str, type: str = "INFO", permission: str = "dashboard.view") -> Dict[str, Any]:
        """Broadcast a notification to users with a specific permission."""
        await self.ensure_auth()
        url = f"{self.base_url}/api/v1/notifications/broadcast"
        payload = {
            "title": title,
            "message": message,
            "type": type,
            "permission": permission
        }
        response = await self.client.post(url, json=payload, headers=self._get_headers())
        response.raise_for_status()
        return response.json()

    async def get_product_performance(self, start_date: str, end_date: str, supplier_name: Optional[str] = None, min_stock: Optional[int] = None) -> list[Dict[str, Any]]:
        """Get product performance analytics."""
        await self.ensure_auth()
        url = f"{self.base_url}/api/v1/reports/product-performance"
        params = {
            "startDate": f"{start_date}T00:00:00Z",
            "endDate": f"{end_date}T23:59:59Z"
        }
        if supplier_name:
            params["supplierName"] = supplier_name
        if min_stock is not None:
            params["minStock"] = min_stock
            
        response = await self.client.get(url, params=params, headers=self._get_headers())
        response.raise_for_status()
        return response.json()

    async def get_supplier_by_name(self, name: str) -> Dict[str, Any]:
        """Get supplier details by name."""
        await self.ensure_auth()
        # URL encode the name
        import urllib.parse
        encoded_name = urllib.parse.quote(name)
        url = f"{self.base_url}/api/v1/suppliers/name/{encoded_name}"
        
        response = await self.client.get(url, headers=self._get_headers())
        if response.status_code == 404:
            return {"error": "Supplier not found"}
        response.raise_for_status()
        return response.json()

    async def get_system_settings(self) -> Dict[str, str]:
        """Fetch public system settings."""
        # This endpoint is public, so auth might not be strictly required,
        # but using _get_headers() won't hurt.
        url = f"{self.base_url}/api/v1/settings/configurations"
        try:
            response = await self.client.get(url, headers=self._get_headers())
            response.raise_for_status()
            return response.json()
        except Exception as e:
            print(f"Failed to fetch settings: {e}")
            return {}

    async def close(self):
        await self.client.aclose()
