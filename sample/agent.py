from dotenv import load_dotenv
from strands import Agent, tool
from strands.models.openai import OpenAIModel

# Load API key and Base URL from .env
load_dotenv()

# Define a custom tool
@tool
def calculate(a: float, b: float, operation: str) -> float:
    """Perform a basic calculation. Operations: add, subtract, multiply, divide."""
    if operation == "add":
        return a + b
    elif operation == "subtract":
        return a - b
    elif operation == "multiply":
        return a * b
    elif operation == "divide":
        return a / b
    return 0.0

# Configure model pointing to your Bedrock Mantle endpoint
model = OpenAIModel(
    model_id="mistral.ministral-3-8b-instruct",  # or mistral.ministral-3-3b-instruct
    client_args={
        "base_url": "https://bedrock-mantle.ap-south-1.api.aws/v1"
    }
)

# Initialize and run the Strands Agent
agent = Agent(model=model, tools=[calculate])

response = agent("What is 125 multiplied by 4?")
print("\nFinal Output:", response)
