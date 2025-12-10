import unittest
from unittest.mock import MagicMock, patch
import json
import sys
import os

# Add parent directory to path to import main
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

# Mock backend_client module BEFORE importing main
sys.modules['backend_client'] = MagicMock()
from main import run_agent, AgentRequest

class TestAgentLoop(unittest.TestCase):
    @patch('main.client')
    @patch('main.backend_client')
    def test_run_agent_loop(self, mock_backend, mock_client):
        # Setup mock LLM responses
        
        # 1. First response: Call tool 'get_product_performance'
        tool_call = MagicMock()
        tool_call.id = 'call_1'
        tool_call.function.name = 'get_product_performance'
        tool_call.function.arguments = json.dumps({"start_date": "2023-01-01", "end_date": "2023-01-31"})
        
        msg1 = MagicMock()
        msg1.tool_calls = [tool_call]
        msg1.content = None
        
        # 2. Second response: Final answer
        msg2 = MagicMock()
        msg2.tool_calls = []
        msg2.content = "The top product is Widget A."
        
        # Mock client.chat.completions.create to return these in sequence
        mock_client.chat.completions.create.side_effect = [
            MagicMock(choices=[MagicMock(message=msg1)]),
            MagicMock(choices=[MagicMock(message=msg2)])
        ]
        
        # Mock backend tool return
        mock_backend.get_product_performance.return_value = [{"product_name": "Widget A", "profit": 100}]
        
        # Run agent
        request = AgentRequest(instruction="Find top product")
        result = run_agent(request)
        
        # Verify result
        self.assertEqual(result["result"], "The top product is Widget A.")
        
        # Verify loop execution
        self.assertEqual(mock_client.chat.completions.create.call_count, 2)
        mock_backend.get_product_performance.assert_called_once()

if __name__ == '__main__':
    unittest.main()
