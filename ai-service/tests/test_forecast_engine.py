import sys
import os
import json
from datetime import datetime, timedelta

# Add parent directory to path
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from analytics import ForecastEngine

def test_forecast_engine():
    engine = ForecastEngine()
    
    # Generate dummy data: increasing trend
    data = []
    base_date = datetime.now() - timedelta(days=90)
    for i in range(90):
        date = base_date + timedelta(days=i)
        quantity = 10 + i + (i % 5) # Linear increase with some noise
        data.append({
            "date": date.strftime("%Y-%m-%d"),
            "quantity": quantity
        })
        
    print("Testing ForecastEngine with increasing trend data...")
    result = engine.predict(data, 30)
    
    print(json.dumps(result, indent=2))
    
    stats = result.get("statistical_forecast", {})
    if stats.get("trend") == "increasing":
        print("PASS: Trend correctly identified as increasing.")
    else:
        print(f"FAIL: Trend identified as {stats.get('trend')}, expected increasing.")
        
    if stats.get("predicted_total_next_30_days") > 0:
        print("PASS: Predicted total is positive.")
    else:
        print("FAIL: Predicted total is not positive.")

if __name__ == "__main__":
    test_forecast_engine()
