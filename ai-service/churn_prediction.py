import os
import json
from typing import Dict, Any, List
from openai import OpenAI
from datetime import datetime

ZAI_API_KEY = os.getenv("ZAI_API_KEY")
ZAI_BASE_URL = os.getenv("ZAI_BASE_URL", "https://api.z.ai/api/paas/v4/")
client = OpenAI(api_key=ZAI_API_KEY, base_url=ZAI_BASE_URL)
MODEL_NAME = "glm-4.5-flash"

class ChurnPredictor:
    def __init__(self):
        self.client = client

    def analyze_churn_risk(self, customer_data: Dict[str, Any], purchase_history: List[Dict[str, Any]]) -> Dict[str, Any]:
        """
        Analyze customer data and purchase history to predict churn risk.
        """
        
        prompt = f"""
        You are an expert CRM analyst. Analyze the following customer data and purchase history to predict the churn risk.
        
        Customer Profile:
        {json.dumps(customer_data, indent=2)}
        
        Recent Purchase History (Last 5 transactions):
        {json.dumps(purchase_history[:5], indent=2)}
        
        Provide a JSON response with the following structure:
        {{
            "churn_risk_score": <float between 0.0 and 1.0, where 1.0 is high risk>,
            "risk_level": <"Low", "Medium", "High">,
            "primary_factors": [<list of strings explaining why>],
            "retention_strategy": <string suggesting a specific action or promotion>,
            "suggested_discount": <integer percentage, e.g. 10 for 10% off, 0 if none needed>
        }}
        
        Consider Recency, Frequency, and Monetary value (RFM) in your analysis.
        If the customer hasn't purchased in a long time compared to their usual frequency, risk is higher.
        """

        try:
            response = self.client.chat.completions.create(
                model=MODEL_NAME,
                messages=[
                    {"role": "system", "content": "You are a helpful CRM AI assistant that outputs JSON."},
                    {"role": "user", "content": prompt}
                ],
                response_format={"type": "json_object"},
                temperature=0.2
            )
            
            content = response.choices[0].message.content
            return json.loads(content)
        except Exception as e:
            print(f"Error generating churn prediction: {e}")
            # Fallback response
            return {
                "churn_risk_score": 0.5,
                "risk_level": "Unknown",
                "primary_factors": ["Error analyzing data"],
                "retention_strategy": "Contact customer support",
                "suggested_discount": 0
            }
