import json
from datetime import datetime, timedelta
from typing import Dict, Any, List
from backend_client import BackendClient
from openai import OpenAI
import os

# Initialize clients
backend_client = BackendClient()
ZAI_API_KEY = os.getenv("ZAI_API_KEY")
ZAI_BASE_URL = os.getenv("ZAI_BASE_URL", "https://api.z.ai/api/paas/v4/")
client = OpenAI(api_key=ZAI_API_KEY, base_url=ZAI_BASE_URL)
MODEL_NAME = "glm-4.5-flash"

def generate_demand_forecast(product_id: int, days_to_forecast: int = 30) -> Dict[str, Any]:
    """
    Generates a demand forecast for a specific product using historical sales data.
    """
    try:
        # 1. Fetch Historical Data (Last 90 days)
        end_date = datetime.now()
        start_date = end_date - timedelta(days=90)
        
        sales_report = backend_client.get_sales_report(
            start_date.strftime("%Y-%m-%d"), 
            end_date.strftime("%Y-%m-%d"),
            product_id=product_id
        )
        
        # 2. Fetch Product Details (for context)
        product_details = backend_client.get_inventory_status(product_id=product_id)
        
        # 3. Prepare Prompt
        sales_trends = sales_report.get("salesTrends", [])
        product_name = product_details.get("name", "Unknown Product")
        current_stock = product_details.get("quantity", 0) # Note: get_inventory_status returns list or dict depending on impl. 
        # Actually get_inventory_status(product_id) returns a single product dict based on backend_client.py
        
        prompt = f"""
        You are an expert Demand Planner. 
        
        Product: {product_name}
        Current Stock: {current_stock}
        
        Historical Sales Data (Last 90 Days):
        {json.dumps(sales_trends, indent=2)}
        
        Task:
        Predict the demand for this product for the next {days_to_forecast} days.
        Consider seasonality, trends, and recent sales velocity.
        
        Output Format (JSON):
        {{
            "predicted_demand": <integer>,
            "confidence_score": <float 0-1>,
            "reasoning": "<concise explanation of the forecast>",
            "daily_forecast": [
                {{"date": "YYYY-MM-DD", "quantity": <integer>}},
                ...
            ]
        }}
        """
        
        # 4. Call LLM
        response = client.chat.completions.create(
            model=MODEL_NAME,
            messages=[
                {"role": "system", "content": "You are a helpful AI assistant for inventory management. Always output valid JSON."},
                {"role": "user", "content": prompt}
            ],
            response_format={"type": "json_object"}
        )
        
        content = response.choices[0].message.content
        forecast_data = json.loads(content)
        
        return forecast_data

    except Exception as e:
        print(f"Error generating forecast: {e}")
        return {"error": str(e)}
