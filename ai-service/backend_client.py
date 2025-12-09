import os
import requests
from typing import Optional, Dict, Any

class BackendClient:
    def __init__(self, base_url: str = "http://localhost:8080"):
        self.base_url = base_url
        self.email = os.getenv("AI_USER_EMAIL", "admin@quantify.com")
        self.password = os.getenv("AI_USER_PASSWORD", "admin123")
        self.token = None
        self._login()

    def _login(self):
        """Authenticate and retrieve a new token."""
        url = f"{self.base_url}/api/v1/users/login"
        try:
            response = requests.post(url, json={"email": self.email, "password": self.password})
            response.raise_for_status()
            data = response.json()
            self.token = data.get("token") # Adjust based on actual login response structure
            if not self.token:
                 # Fallback if token is nested or named differently
                 self.token = data.get("accessToken")
            print(f"AI Agent authenticated as {self.email}")
        except Exception as e:
            print(f"Failed to authenticate AI Agent: {e}")
            # Fallback to manual token if login fails (e.g. dev mode)
            self.token = os.getenv("BACKEND_API_TOKEN")

    def _get_headers(self) -> Dict[str, str]:
        headers = {"Content-Type": "application/json"}
        if self.token:
            headers["Authorization"] = f"Bearer {self.token}"
        return headers

    def get_inventory_status(self, product_id: Optional[int] = None) -> Dict[str, Any]:
        """Fetch inventory status for a product or all products."""
        url = f"{self.base_url}/api/v1/products"
        if product_id:
            url = f"{self.base_url}/api/v1/products/{product_id}"
        
        response = requests.get(url, headers=self._get_headers())
        response.raise_for_status()
        return response.json()

    def create_purchase_order(self, product_id: int, quantity: int, supplier_id: int) -> Dict[str, Any]:
        """Create a draft purchase order."""
        url = f"{self.base_url}/api/v1/replenishment/purchase-orders"
        
        # We need to fetch product price first to populate unit price
        # For now, let's assume we can fetch it or just pass 0 and let backend handle/validate? 
        # The backend requires UnitPrice.
        # Let's fetch product details first.
        product_url = f"{self.base_url}/api/v1/products/{product_id}"
        prod_resp = requests.get(product_url, headers=self._get_headers())
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
        
        response = requests.post(url, json=payload, headers=self._get_headers())
        response.raise_for_status()
        return response.json()

    def get_sales_report(self, start_date: str, end_date: str) -> Dict[str, Any]:
        """Fetch sales report."""
        url = f"{self.base_url}/api/v1/reports/sales-trends" # Correct endpoint
        payload = {"startDate": start_date, "endDate": end_date, "interval": "daily"}
        response = requests.post(url, json=payload, headers=self._get_headers())
        response.raise_for_status()
        return response.json()
