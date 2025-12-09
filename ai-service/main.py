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

# Agent Implementation
from backend_client import BackendClient
backend_client = BackendClient()

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

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8001)
