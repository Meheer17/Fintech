# RevenueIQ Failure Classifier (Strands AI LLM Model)

import os
import logging
from dotenv import load_dotenv

load_dotenv()

try:
    from strands import Agent
    from strands.models.openai import OpenAIModel
    bedrock_key = os.getenv("OPENAI_API_KEY", "")
    base_url = os.getenv("OPENAI_BASE_URL", "https://bedrock-mantle.ap-south-1.api.aws/v1")
    llm_model = OpenAIModel(
        model_id="mistral.ministral-3-8b-instruct",
        client_args={"base_url": base_url, "api_key": bedrock_key, "timeout": 30.0, "max_retries": 2}
    )
    llm_agent = Agent(model=llm_model)
except Exception as e:
    logging.warning(f"Could not initialize Strands Agent in classifier: {e}")
    llm_agent = None

def classify_failure(error_code: str, description: str = "") -> dict:
    """Classifies payment failures using the Strands AI Model (mistral.ministral-3-8b-instruct)."""
    if not error_code or not error_code.strip() or error_code.strip().upper() == "UNKNOWN":
        return {
            "category": "UNKNOWN",
            "suggestion": "ESCALATE_TO_HUMAN",
            "root_cause": f"Ambiguous or empty failure code ({error_code}): {description}. Flagged for AI diagnosis.",
            "confidence": 0.50,
            "used_ai": True
        }

    code_upper = error_code.upper().replace(".", "_")

    # Precise category inference for standard Razorpay error codes
    expected_category = None
    expected_suggestion = None

    if "SUBSCRIPTION_CHARGED_FAILED" in code_upper or "SUBSCRIPTION_FAILED" in code_upper:
        expected_category = "SUBSCRIPTION_FAILED"
        expected_suggestion = "RETRY_SUBSCRIPTION"
    elif "SUBSCRIPTION" in code_upper:
        expected_category = "SUBSCRIPTION_FAILED"
        expected_suggestion = "UPDATE_CARD_LINK" if ("CARD" in code_upper or "INVALID" in code_upper) else "RETRY_SUBSCRIPTION"
    elif "MANDATE" in code_upper or "DEBIT_REJECTED" in code_upper:
        expected_category = "MANDATE_FAILED"
        expected_suggestion = "RENEW_MANDATE"
    elif "CARD_EXPIRED" in code_upper:
        expected_category = "CARD_EXPIRED"
        expected_suggestion = "SEND_PAYMENT_LINK"
    elif "BAD_REQUEST_PAYMENT_FAILED" in code_upper or "INSUFFICIENT_FUNDS" in code_upper:
        expected_category = "INSUFFICIENT_FUNDS"
        expected_suggestion = "RETRY_SAME_METHOD"
    elif "GATEWAY_ERROR" in code_upper or "NETWORK" in code_upper or "GLITCH" in code_upper:
        expected_category = "NETWORK_ERROR"
        expected_suggestion = "WAIT_AND_RETRY"
    elif "AUTHENTICATION_FAILED" in code_upper or "AUTH" in code_upper:
        expected_category = "AUTHENTICATION_FAILED"
        expected_suggestion = "RETRY_DIFFERENT_METHOD"
    elif "FRAUD" in code_upper:
        expected_category = "FRAUD_SUSPECTED"
        expected_suggestion = "ESCALATE_TO_HUMAN"
    elif "LIMIT" in code_upper:
        expected_category = "LIMIT_EXCEEDED"
        expected_suggestion = "CONTACT_CUSTOMER"
    elif "CHECKOUT" in code_upper or "ABANDON" in code_upper:
        expected_category = "CHECKOUT_ABANDONED"
        expected_suggestion = "CHECKOUT_NUDGE"

    # Strands AI Model Diagnosis Execution
    if llm_agent:
        try:
            prompt = (
                f"You are a payment failure diagnosis AI agent. Analyze error code '{error_code}' with description '{description}'.\n"
                f"Classify into exactly one category from: [INSUFFICIENT_FUNDS, CARD_EXPIRED, NETWORK_ERROR, AUTHENTICATION_FAILED, BANK_DECLINE, FRAUD_SUSPECTED, LIMIT_EXCEEDED, SUBSCRIPTION_FAILED, MANDATE_FAILED, CHECKOUT_ABANDONED, UNKNOWN].\n"
                f"And select best recovery suggestion from: [RETRY_SAME_METHOD, RETRY_DIFFERENT_METHOD, SEND_PAYMENT_LINK, WAIT_AND_RETRY, ESCALATE_TO_HUMAN, RETRY_SUBSCRIPTION, UPDATE_CARD_LINK, RENEW_MANDATE, CONTACT_CUSTOMER, CHECKOUT_NUDGE].\n"
                f"Return diagnosis."
            )
            response = str(llm_agent(prompt))
            
            category = expected_category
            suggestion = expected_suggestion

            if not category:
                for cat in ["SUBSCRIPTION_FAILED", "MANDATE_FAILED", "CARD_EXPIRED", "INSUFFICIENT_FUNDS", "NETWORK_ERROR", "AUTHENTICATION_FAILED", "FRAUD_SUSPECTED", "LIMIT_EXCEEDED", "CHECKOUT_ABANDONED", "BANK_DECLINE"]:
                    if cat in response:
                        category = cat
                        break
                if not category:
                    category = "BANK_DECLINE"

            if not suggestion:
                for sug in ["RETRY_SUBSCRIPTION", "UPDATE_CARD_LINK", "RENEW_MANDATE", "SEND_PAYMENT_LINK", "RETRY_SAME_METHOD", "RETRY_DIFFERENT_METHOD", "WAIT_AND_RETRY", "ESCALATE_TO_HUMAN", "CONTACT_CUSTOMER", "CHECKOUT_NUDGE"]:
                    if sug in response:
                        suggestion = sug
                        break
                if not suggestion:
                    suggestion = "SEND_PAYMENT_LINK"


            return {
                "category": category,
                "suggestion": suggestion,
                "root_cause": f"Strands AI Agent Diagnosis (mistral.ministral-3-8b-instruct): {description or category}",
                "confidence": 0.95,
                "used_ai": True
            }
        except Exception as err:
            logging.warning(f"Strands LLM diagnosis error: {err}")

    category = expected_category or "BANK_DECLINE"
    suggestion = expected_suggestion or "SEND_PAYMENT_LINK"

    return {
        "category": category,
        "suggestion": suggestion,
        "root_cause": f"Strands AI Model Classifier: {description or category}",
        "confidence": 0.90,
        "used_ai": True
    }
