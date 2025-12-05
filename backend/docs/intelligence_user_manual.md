# 🧠 Intelligence Module - User Manual

The **Intelligence** module in Quantify is your command center for data-driven decision making. It uses your historical sales and stock data to provide accurate forecasts, automated reorder suggestions, and detailed performance reports.

---

## 1. Reorder Suggestions

This section automates your replenishment process by identifying products that are running low.

### How it Works
The system constantly monitors your inventory. A "Suggestion" is generated when:
1.  **Stock is Low**: A product's `Current Stock` falls at or below its configured `Low Stock Level`.
2.  **No Pending Orders**: There are no existing pending suggestions or active Purchase Orders for this product (to prevent double-ordering).
3.  **Calculation**: The system calculates the `Suggested Quantity` needed to bring the stock back up to the `Over Stock Level` (Target).

### How to Use
-   **View Suggestions**: The table lists all products that need attention. You can see the Supplier, Current Stock, and Suggested Quantity.
-   **Refresh List**: Click the **"Refresh"** button in the top-right of the card. This forces the system to re-scan your entire inventory and generate new suggestions based on the very latest data.
-   **Create PO**: Click the **"Create PO"** button next to a suggestion to instantly generate a draft Purchase Order.

---

## 2. Demand Forecasting

Predict future demand for specific products to plan ahead.

### How to Use
1.  **Select Product**: Enter the Product ID you want to analyze. Leave it blank to run a general analysis (backend optimization).
2.  **Select Period**: Choose the lookback period (e.g., "30 Days"). The system will analyze sales from this period to predict future trends.
3.  **Generate**: Click **"Generate Forecast"**.
4.  **Result**: The system will display the predicted demand. This uses a **time-weighted algorithm**, meaning recent sales have more impact on the prediction than older ones.

---

## 3. Performance Reports

Generate deep insights into your business performance.

### Available Reports
-   **Sales Trends**: Visualizes your sales performance over time.
    -   *Features*: Includes a dynamic **sparkline chart** that shows the ups and downs of your sales volume at a glance.
-   **Inventory Turnover**: Measures how efficiently you are selling your stock. High turnover means efficient management; low turnover might indicate overstocking.
-   **Profit Margin**: Calculates your gross profit (Revenue - Cost) and margin percentage.

### How to Run a Report
1.  **Select Date Range**: Use the date picker at the top of the Reports card to define the period you want to analyze.
2.  **Choose Report**: Click on the tab for the report you want (Sales, Turnover, or Margin).
3.  **Run**: Click the **"Run Report"** button. The data (and charts) will update instantly.
4.  **Export**: Use the download icon to export the data for further analysis in Excel or other tools.

---

## 💡 Tips for Success
-   **Set Accurate Alerts**: The "Reorder Suggestions" are only as good as your settings. Ensure every product has a realistic `Low Stock Level` and `Over Stock Level` set in the Alerts configuration.
-   **Regular Refreshes**: Click "Refresh" on the suggestions list before placing weekly orders to ensure you catch everything.
