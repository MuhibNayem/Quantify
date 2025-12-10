import pandas as pd
from sklearn.linear_model import LinearRegression
import numpy as np
from typing import List, Dict, Any

class ForecastEngine:
    def __init__(self):
        pass

    def predict(self, sales_data: List[Dict[str, Any]], days_to_forecast: int = 30) -> Dict[str, Any]:
        """
        Generates a statistical forecast using Linear Regression.
        
        Args:
            sales_data: List of dicts with 'date' (YYYY-MM-DD) and 'quantity'.
            days_to_forecast: Number of days to predict.
            
        Returns:
            Dict containing:
            - trend: 'up', 'down', or 'stable'
            - predicted_total: Total predicted quantity for the period.
            - daily_average: Average daily sales.
            - confidence_interval: Simple variance-based confidence.
            - forecast_points: List of predicted values.
        """
        if not sales_data:
            return {
                "error": "No historical data provided"
            }

        df = pd.DataFrame(sales_data)
        df['date'] = pd.to_datetime(df['date'])
        df = df.sort_values('date')
        
        # Fill missing dates with 0 sales
        idx = pd.date_range(df['date'].min(), df['date'].max())
        df = df.set_index('date').reindex(idx, fill_value=0).reset_index()
        df.rename(columns={'index': 'date'}, inplace=True)
        
        # Prepare data for regression
        df['day_ordinal'] = df['date'].map(pd.Timestamp.toordinal)
        
        X = df[['day_ordinal']]
        y = df['quantity']
        
        model = LinearRegression()
        model.fit(X, y)
        
        # Predict future
        last_date = df['date'].max()
        future_dates = [last_date + pd.Timedelta(days=x) for x in range(1, days_to_forecast + 1)]
        future_ordinals = [[d.toordinal()] for d in future_dates]
        
        predictions = model.predict(future_ordinals)
        predictions = np.maximum(predictions, 0) # No negative sales
        
        predicted_total = int(np.sum(predictions))
        daily_average = float(np.mean(y))
        
        # Determine trend
        slope = model.coef_[0]
        if slope > 0.1:
            trend = "increasing"
        elif slope < -0.1:
            trend = "decreasing"
        else:
            trend = "stable"
            
        return {
            "statistical_forecast": {
                "trend": trend,
                "slope": float(slope),
                "predicted_total_next_30_days": predicted_total,
                "historical_daily_average": daily_average,
                "method": "Linear Regression (scikit-learn)"
            }
        }
