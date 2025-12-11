import os
import json
import asyncio
from contextlib import asynccontextmanager
from fastapi import FastAPI, HTTPException, Body, Request
from pydantic import BaseModel
from openai import AsyncOpenAI
from dotenv import load_dotenv
from apscheduler.schedulers.asyncio import AsyncIOScheduler
from apscheduler.triggers.cron import CronTrigger

# Load environment variables
load_dotenv()

# Configuration
ZAI_API_KEY = os.getenv("ZAI_API_KEY")
ZAI_BASE_URL = os.getenv("ZAI_BASE_URL", "https://api.z.ai/api/paas/v4/")

if not ZAI_API_KEY:
    print("Warning: ZAI_API_KEY is not set.")

# Initialize OpenAI Client (Global)
client = AsyncOpenAI(
    api_key=ZAI_API_KEY,
    base_url=ZAI_BASE_URL
)

from backend_client import BackendClient
backend_client = BackendClient()

MODEL_NAME = "glm-4.5-flash"
scheduler = AsyncIOScheduler()

@asynccontextmanager
async def lifespan(app: FastAPI):
    # Startup
    print("AI Service Starting up...")
    
    # Initialize Scheduler
    wake_up_time = await get_system_setting("ai_wake_up_time", "07:00")
    try:
        hour, minute = wake_up_time.split(":")
    except ValueError:
        hour, minute = "07", "00"
        
    trigger = CronTrigger(hour=hour, minute=minute)
    scheduler.add_job(daily_morning_check, trigger, id="daily_morning_check", replace_existing=True)
    scheduler.start()
    print(f"Scheduler started. Daily Morning Check set for {wake_up_time}")
    
    yield
    
    # Shutdown
    print("AI Service Shutting down...")
    scheduler.shutdown()
    await backend_client.close()
    await client.close()

app = FastAPI(title="Quantify AI Service", lifespan=lifespan)

class InsightRequest(BaseModel):
    query: str

class ForecastAnalysisRequest(BaseModel):
    dashboard_data: dict

@app.get("/health")
async def health_check():
    return {"status": "ok"}

@app.post("/insight")
async def generate_business_insight(request: InsightRequest):
    try:
        response = await client.chat.completions.create(
            model=MODEL_NAME,
            messages=[
                {"role": "system", "content": "You are an expert business analyst for a retail chain. Provide concise, actionable insights."},
                {"role": "user", "content": request.query}
            ],
            extra_body={
                "thinking": {
                    "type": "enabled"
                }
            }
        )
        
        content = response.choices[0].message.content
        return {"insight": content}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.post("/analyze-forecast")
async def analyze_forecast(request: ForecastAnalysisRequest):
    try:
        data_str = json.dumps(request.dashboard_data, indent=2)
        
        prompt = f"""
Analyze the following inventory forecast data and provide a strategic summary.
Highlight critical low stock items and suggest action plans.

Data:
{data_str}
"""

        response = await client.chat.completions.create(
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

class ForecastRequest(BaseModel):
    product_id: int
    period_days: int = 30

@app.post("/forecast")
async def generate_forecast(request: ForecastRequest):
    return await generate_demand_forecast(request.product_id, request.period_days)

class ChurnAnalysisRequest(BaseModel):
    customer_data: dict
    purchase_history: list[dict]

# Note: Assuming ChurnPredictor needs to be made async too if it uses backend_client
# For now, we'll assume it's CPU bound or wrapper.
from churn_prediction import ChurnPredictor

@app.post("/predict-churn")
async def predict_churn(request: ChurnAnalysisRequest):
    predictor = ChurnPredictor()
    # If analyze_churn_risk is synchronous, run it in thread pool to avoid blocking
    # result = await asyncio.to_thread(predictor.analyze_churn_risk, request.customer_data, request.purchase_history)
    # But if it uses backend_client internally, it needs to be updated.
    # Let's assume for now we call it directly, if it blocks, we fix next.
    result = predictor.analyze_churn_risk(request.customer_data, request.purchase_history)
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
    },
    {
        "type": "function",
        "function": {
            "name": "get_product_performance",
            "description": "Get product performance analytics (profit, revenue, stock coverage). Useful for finding top products or low stock items.",
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
                    },
                    "supplier_name": {
                        "type": "string",
                        "description": "Filter by supplier name (optional)."
                    },
                    "min_stock": {
                        "type": "integer",
                        "description": "Filter by minimum stock level (optional)."
                    }
                },
                "required": ["start_date", "end_date"]
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "get_supplier_by_name",
            "description": "Get supplier details by name. Useful for finding supplier ID.",
            "parameters": {
                "type": "object",
                "properties": {
                    "name": {
                        "type": "string",
                        "description": "Name of the supplier."
                    }
                },
                "required": ["name"]
            }
        }
    }
]

class AgentRequest(BaseModel):
    instruction: str

@app.post("/agent/run")
async def run_agent(request: AgentRequest):
    messages = [
        {"role": "system", "content": "You are an autonomous agent managing a retail inventory system. Use the available tools to fulfill the user's request. You can chain multiple tools to achieve complex goals. Always verify you have the necessary IDs (like supplier_id) before creating orders."},
        {"role": "user", "content": request.instruction}
    ]

    max_iterations = 5
    
    try:
        for i in range(max_iterations):
            # Call LLM
            response = await client.chat.completions.create(
                model=MODEL_NAME,
                messages=messages,
                tools=tools,
                tool_choice="auto"
            )

            response_message = response.choices[0].message
            tool_calls = response_message.tool_calls
            
            # Append assistant's response (thought/tool call)
            messages.append(response_message)

            if not tool_calls:
                # No more tools to call, we are done
                return {"result": response_message.content}
            
            # Execute tools
            for tool_call in tool_calls:
                function_name = tool_call.function.name
                function_args = json.loads(tool_call.function.arguments)
                
                tool_output = None
                try:
                    if function_name == "get_inventory_status":
                        tool_output = await backend_client.get_inventory_status(**function_args)
                    elif function_name == "create_purchase_order":
                        tool_output = await backend_client.create_purchase_order(**function_args)
                    elif function_name == "get_sales_report":
                        tool_output = await backend_client.get_sales_report(**function_args)
                    elif function_name == "get_product_performance":
                        tool_output = await backend_client.get_product_performance(**function_args)
                    elif function_name == "get_supplier_by_name":
                        tool_output = await backend_client.get_supplier_by_name(**function_args)
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
        
        return {"result": "Agent reached maximum iteration limit without final answer."}

    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

async def get_system_setting(key: str, default: str) -> str:
    """Fetch a system setting from the backend."""
    configs = await backend_client.get_system_settings()
    return configs.get(key, default)

async def daily_morning_check():
    """
    Proactive Routine:
    1. Trigger backend alert check (uses existing ProductAlertSettings).
    2. Fetch active alerts.
    3. Generate a briefing and draft POs for critical alerts.
    """
    print("Running Daily Morning Check...")
    try:
        # 1. Trigger Alert Check
        await backend_client.trigger_alert_check()
        
        # 2. Fetch Active Alerts
        alerts = await backend_client.get_active_alerts()
        
        # 3. Fetch Yesterday's Sales
        from datetime import datetime, timedelta
        yesterday = (datetime.now() - timedelta(days=1)).strftime('%Y-%m-%d')
        sales_report = await backend_client.get_sales_report(yesterday, yesterday)
        
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
        
        response = await client.chat.completions.create(
            model=MODEL_NAME,
            messages=[
                {"role": "system", "content": "You are a proactive retail assistant."},
                {"role": "user", "content": prompt}
            ]
        )
        
        briefing = response.choices[0].message.content
        print(f"Morning Briefing:\n{briefing}")
        
        # Broadcast the briefing as a notification
        await backend_client.broadcast_notification(
            title="Daily Morning Briefing",
            message=briefing,
            type="INFO",
            permission="dashboard.view"
        )
        print("Morning Briefing broadcasted to dashboard users.")
        
    except Exception as e:
        print(f"Error in Daily Morning Check: {e}")

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8001)
if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8001)
