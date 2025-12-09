import os
import json
from fastapi import FastAPI, HTTPException, Body
from pydantic import BaseModel
from openai import OpenAI
from dotenv import load_dotenv

# Load environment variables
load_dotenv()

app = FastAPI(title="Quantify AI Service")

# Configuration
ZAI_API_KEY = os.getenv("ZAI_API_KEY")
ZAI_BASE_URL = os.getenv("ZAI_BASE_URL", "https://api.z.ai/api/paas/v4/")

if not ZAI_API_KEY:
    print("Warning: ZAI_API_KEY is not set.")

# Initialize OpenAI Client
client = OpenAI(
    api_key=ZAI_API_KEY,
    base_url=ZAI_BASE_URL
)

MODEL_NAME = "glm-4.5-flash"

class InsightRequest(BaseModel):
    query: str

class ForecastAnalysisRequest(BaseModel):
    dashboard_data: dict

@app.get("/health")
def health_check():
    return {"status": "ok"}

@app.post("/insight")
def generate_business_insight(request: InsightRequest):
    try:
        response = client.chat.completions.create(
            model=MODEL_NAME,
            messages=[
                {"role": "system", "content": "You are an expert business analyst for a retail chain. Provide concise, actionable insights."},
                {"role": "user", "content": request.query}
            ],
            # Enable thinking mode for deeper reasoning if needed
            extra_body={
                "thinking": {
                    "type": "enabled"
                }
            }
        )
        
        # Extract content and reasoning (if available)
        content = response.choices[0].message.content
        # Note: The doc shows reasoning_content in stream delta. For non-stream, it might be in message or extra fields.
        # We'll just return the content for now.
        
        return {"insight": content}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.post("/analyze-forecast")
def analyze_forecast(request: ForecastAnalysisRequest):
    try:
        # Convert dashboard data to string for the prompt
        data_str = json.dumps(request.dashboard_data, indent=2)
        
        prompt = f"""
Analyze the following inventory forecast data and provide a strategic summary.
Highlight critical low stock items and suggest action plans.

Data:
{data_str}
"""

        response = client.chat.completions.create(
            model=MODEL_NAME,
            messages=[
                {"role": "system", "content": "You are an inventory management expert. Analyze the data and provide a strategic summary."},
                {"role": "user", "content": prompt}
            ],
            extra_body={
                "thinking": {
                    "type": "enabled"
                }
            }
        )
        
        content = response.choices[0].message.content
        return {"analysis": content}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

from forecasting import generate_demand_forecast
from forecasting import generate_demand_forecast, DemandForecaster

class ForecastRequest(BaseModel):
    product_id: int
    days: int = 30

@app.post("/forecast")
async def generate_forecast(request: Request):
    data = await request.json()
    product_id = data.get("product_id")
    period_days = data.get("period_days", 30)
    
    # Fetch sales history from backend
    # For now, we'll mock or fetch if we had the client set up for it
    # In a real scenario, we'd fetch sales history here
    
    # Mock sales history for now to test the flow
    sales_history = [
        {"date": "2023-01-01", "quantity": 10},
        {"date": "2023-01-02", "quantity": 12},
        {"date": "2023-01-03", "quantity": 15},
        # ... more data
    ]
    
    forecaster = DemandForecaster()
    result = forecaster.generate_forecast(sales_history, period_days)
    return result

# Agent Implementation
from backend_client import BackendClient
from churn_prediction import ChurnPredictor

@app.post("/predict-churn")
async def predict_churn(request: Request):
    data = await request.json()
    customer_data = data.get("customer_data")
    purchase_history = data.get("purchase_history")
    
    predictor = ChurnPredictor()
    result = predictor.analyze_churn_risk(customer_data, purchase_history)
    return result

tools = [
    {
        "type": "function",
        "function": {
            "name": "get_inventory_status",
            "description": "Get inventory status for a specific product or all products.",
            "parameters": {
                "type": "object",
                "properties": {
                    "product_id": {
                        "type": "integer",
                        "description": "The ID of the product to check. If omitted, returns all products."
                    }
                },
                "required": []
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "create_purchase_order",
            "description": "Create a new purchase order for a product.",
            "parameters": {
                "type": "object",
                "properties": {
                    "product_id": {
                        "type": "integer",
                        "description": "The ID of the product to order."
                    },
                    "quantity": {
                        "type": "integer",
                        "description": "The quantity to order."
                    },
                    "supplier_id": {
                        "type": "integer",
                        "description": "The ID of the supplier."
                    }
                },
                "required": ["product_id", "quantity", "supplier_id"]
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "get_sales_report",
            "description": "Get sales report for a specific date range.",
            "parameters": {
                "type": "object",
                "properties": {
                    "start_date": {
                        "type": "string",
                        "description": "Start date in YYYY-MM-DD format."
                    },
                    "end_date": {
                        "type": "string",
                        "description": "End date in YYYY-MM-DD format."
                    }
                },
                "required": ["start_date", "end_date"]
            }
        }
    }
]

class AgentRequest(BaseModel):
    instruction: str

@app.post("/agent/run")
def run_agent(request: AgentRequest):
    messages = [
        {"role": "system", "content": "You are an autonomous agent managing a retail inventory system. Use the available tools to fulfill the user's request."},
        {"role": "user", "content": request.instruction}
    ]

    try:
        # First call to LLM
        response = client.chat.completions.create(
            model=MODEL_NAME,
            messages=messages,
            tools=tools,
            tool_choice="auto"
        )

        response_message = response.choices[0].message
        tool_calls = response_message.tool_calls

        if tool_calls:
            messages.append(response_message)
            
            for tool_call in tool_calls:
                function_name = tool_call.function.name
                function_args = json.loads(tool_call.function.arguments)
                
                tool_output = None
                try:
                    if function_name == "get_inventory_status":
                        tool_output = backend_client.get_inventory_status(**function_args)
                    elif function_name == "create_purchase_order":
                        tool_output = backend_client.create_purchase_order(**function_args)
                    elif function_name == "get_sales_report":
                        tool_output = backend_client.get_sales_report(**function_args)
                    else:
                        tool_output = {"error": "Unknown function"}
                except Exception as e:
                    tool_output = {"error": str(e)}
                
                messages.append({
                    "tool_call_id": tool_call.id,
                    "role": "tool",
                    "name": function_name,
                    "content": json.dumps(tool_output)
                })
                
            # Second call to LLM to summarize
            final_response = client.chat.completions.create(
                model=MODEL_NAME,
                messages=messages
            )
            return {"result": final_response.choices[0].message.content}
        
        return {"result": response_message.content}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

# Scheduler Implementation
from apscheduler.schedulers.background import BackgroundScheduler
from apscheduler.triggers.cron import CronTrigger
import requests

scheduler = BackgroundScheduler()

def get_system_setting(key: str, default: str) -> str:
    """Fetch a system setting from the backend."""
    try:
        # We need to authenticate. For now, we'll use the backend_client's token if available,
        # or just assume internal network trust for settings if public.
        # Actually, settings/configurations is public.
        response = requests.get(f"http://localhost:8080/api/v1/settings/configurations")
        if response.status_code == 200:
            configs = response.json()
            return configs.get(key, default)
    except Exception as e:
        print(f"Failed to fetch setting {key}: {e}")
    return default

def daily_morning_check():
    """
    Proactive Routine:
    1. Trigger backend alert check (uses existing ProductAlertSettings).
    2. Fetch active alerts.
    3. Generate a briefing and draft POs for critical alerts.
    """
    print("Running Daily Morning Check...")
    try:
        # 1. Trigger Alert Check
        backend_client.trigger_alert_check()
        
        # 2. Fetch Active Alerts
        alerts = backend_client.get_active_alerts()
        
        # 3. Fetch Yesterday's Sales
        from datetime import datetime, timedelta
        yesterday = (datetime.now() - timedelta(days=1)).strftime('%Y-%m-%d')
        sales_report = backend_client.get_sales_report(yesterday, yesterday)
        
        if not alerts and not sales_report:
            print("No active alerts and no sales data. Inventory is healthy.")
            return

        # 4. Analyze with AI
        prompt = f"""
        You are the Proactive AI Agent. It is the Daily Morning Check.
        
        Yesterday's Sales ({yesterday}):
        {json.dumps(sales_report, indent=2)}
        
        Active Alerts:
        {json.dumps(alerts, indent=2)}
        
        Task:
        1. Provide a quick summary of yesterday's sales performance.
        2. Summarize any critical inventory alerts.
        3. Recommend actions (e.g., "Draft PO for Product X").
        
        Output a concise morning briefing.
        """
        
        response = client.chat.completions.create(
            model=MODEL_NAME,
            messages=[
                {"role": "system", "content": "You are a proactive retail assistant."},
                {"role": "user", "content": prompt}
            ]
        )
        
        briefing = response.choices[0].message.content
        print(f"Morning Briefing:\n{briefing}")
        
        # Broadcast the briefing as a notification
        backend_client.broadcast_notification(
            title="Daily Morning Briefing",
            message=briefing,
            type="INFO",
            permission="dashboard.view"
        )
        print("Morning Briefing broadcasted to dashboard users.")
        
    except Exception as e:
        print(f"Error in Daily Morning Check: {e}")

@app.on_event("startup")
def start_scheduler():
    # Fetch wake up time from settings, default to 07:00
    wake_up_time = get_system_setting("ai_wake_up_time", "07:00")
    hour, minute = wake_up_time.split(":")
    
    trigger = CronTrigger(hour=hour, minute=minute)
    scheduler.add_job(daily_morning_check, trigger, id="daily_morning_check", replace_existing=True)
    scheduler.start()
    print(f"Scheduler started. Daily Morning Check set for {wake_up_time}")

@app.on_event("shutdown")
def stop_scheduler():
    scheduler.shutdown()

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8001)
