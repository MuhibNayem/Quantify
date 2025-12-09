# Quantify AI Agent: Shop Owner's Manual

## 1. Introduction
Welcome to your new **AI Co-Pilot**. This isn't just a chatbot; it's a proactive member of your team that works 24/7 to keep your inventory healthy and your sales growing.

### What can it do?
*   **Monitor Inventory**: Watches stock levels in real-time.
*   **Predict Issues**: Warns you *before* you run out of stock.
*   **Draft Orders**: Autonomously prepares Purchase Orders for your approval.
*   **Analyze Business**: Answers questions like "How are sales this week?" or "Who is my best employee?".

---

## 2. Getting Started

### A. Configuration
The AI needs to know when to start its day.
1.  Go to **Settings > System Configuration**.
2.  Find the setting: `ai_wake_up_time`.
3.  Set it to your preferred time (e.g., `07:00` or `08:30`).
    *   *Default: 07:00 AM*

### B. Setting Thresholds
The AI uses your product settings to decide when to reorder.
1.  Go to **Inventory > Products**.
2.  Select a product and go to the **Alerts** tab.
3.  Set the **Low Stock Level** (e.g., 10 units).
    *   *The AI will trigger an alert when stock hits this number.*

---

## 3. Daily Workflow: A Day with Your AI

### 🌅 07:00 AM: The Morning Briefing
While you are having coffee, the AI wakes up and runs its **Daily Morning Check**.
1.  **Scans Inventory**: Checks every single item against your thresholds.
2.  **Checks Expiry**: Looks for batches expiring soon.
3.  **Analyzes Sales**: Reviews yesterday's performance.
4.  **Action**: It sends you a **Morning Briefing** (via Notification/Slack).
    *   *"Good morning! Sales were up 5% yesterday. However, we are low on 'Wireless Mouse' and 'HDMI Cables'. I have drafted POs for them."*

### 🏢 During the Day: Your Assistant
You can ask the AI questions at any time via the **AI Insight** dashboard.
*   **You**: "Check stock for iPhone 13."
*   **AI**: "We have 15 units in the Main Warehouse. Sales are trending up, so you might want to reorder soon."
*   **You**: "Create a PO for 50 units."
*   **AI**: "Done. PO #1024 has been created and sent to 'Apple Inc.' for approval."

### 🌙 Overnight: Self-Healing
The AI continues to monitor stock. If a sudden rush depletes an item to critical levels (e.g., 0 units), the AI can **autonomously** draft an emergency PO to ensure you don't lose sales tomorrow.

---

## 4. Advanced Features

### 🧠 Proactive Autonomy
The AI doesn't just wait for you. It actively looks for problems.
*   **Dead Stock Detection**: Identifies items that haven't sold in 90 days and suggests a discount.
*   **Supplier Rating**: Tells you which suppliers are slow or often send damaged goods.

### 📊 Strategic Analysis
Ask complex questions:
*   *"What should I put on sale this weekend?"*
*   *"Which category has the highest profit margin?"*
*   *"Who are my top 10 customers at risk of churning?"*

---

## 5. Troubleshooting

**Q: The AI didn't run this morning.**
*   Check if the **Backend** and **AI Service** are running.
*   Verify the `ai_wake_up_time` in Settings.

**Q: The AI is ordering too much.**
*   Check your **Overstock Level** settings for that product.
*   The AI respects the limits you set.

**Q: I don't see the Morning Briefing.**
*   Ensure your **Notification Settings** (Email/SMS) are enabled in your User Profile.

---

*Quantify AI: Empowering your business with intelligence.*
