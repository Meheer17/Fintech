# RevenueIQ Failure Classifier (Rule Engine + Strands LLM Fallback)

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

ERROR_CODE_MAP = {
    "BAD_REQUEST_PAYMENT_FAILED": ("INSUFFICIENT_FUNDS", "RETRY_SAME_METHOD"),
    "GATEWAY_ERROR": ("NETWORK_ERROR", "WAIT_AND_RETRY"),
    "CARD_EXPIRED": ("CARD_EXPIRED", "SEND_PAYMENT_LINK"),
    "AUTHENTICATION_FAILED": ("AUTHENTICATION_FAILED", "RETRY_DIFFERENT_METHOD"),
    "BANK_TECHNICAL_GLITCH": ("BANK_DECLINE", "WAIT_AND_RETRY"),
    "FRAUD_SUSPECTED": ("FRAUD_SUSPECTED", "ESCALATE_TO_HUMAN"),
    "LIMIT_EXCEEDED": ("LIMIT_EXCEEDED", "CONTACT_CUSTOMER"),
    "SUBSCRIPTION.CHARGED.FAILED": ("SUBSCRIPTION_FAILED", "RETRY_SUBSCRIPTION"),
    "SUBSCRIPTION_CHARGED_FAILED": ("SUBSCRIPTION_FAILED", "RETRY_SUBSCRIPTION"),
    "SUBSCRIPTION_CARD_INVALID": ("SUBSCRIPTION_FAILED", "UPDATE_CARD_LINK"),
    "MANDATE_EXPIRED": ("MANDATE_FAILED", "RENEW_MANDATE"),
    "DEBIT_REJECTED": ("MANDATE_FAILED", "RENEW_MANDATE"),
    "MANDATE_NOT_ACTIVE": ("MANDATE_FAILED", "RENEW_MANDATE"),
    "INSUFFICIENT_BALANCE_MANDATE": ("MANDATE_FAILED", "RENEW_MANDATE"),
    "CHECKOUT_ABANDONED": ("CHECKOUT_ABANDONED", "CHECKOUT_NUDGE"),
}

def classify_failure(error_code: str, description: str = "") -> dict:
    if not error_code or not error_code.strip() or error_code.strip().upper() == "UNKNOWN":
        return {
            "category": "UNKNOWN",
            "suggestion": "ESCALATE_TO_HUMAN",
            "root_cause": f"Ambiguous or empty failure code ({error_code}): {description}. Flagged for AI/human diagnosis.",
            "confidence": 0.50,
            "used_ai": False
        }

    code_upper = error_code.upper()
    code_norm = code_upper.replace(".", "_")
    for key, (category, suggestion) in ERROR_CODE_MAP.items():
        key_norm = key.replace(".", "_")
        if key in code_upper or key_norm in code_norm:
            return {
                "category": category,
                "suggestion": suggestion,
                "root_cause": f"Diagnosed by rule engine: {description or category}",
                "confidence": 0.95,
                "used_ai": False
            }
    
    if llm_agent:
        try:
            prompt = (
                f"Diagnose payment failure code '{error_code}' with description '{description}'. "
                f"Classify into category (INSUFFICIENT_FUNDS, BANK_DECLINE, CARD_EXPIRED, NETWORK_ERROR, AUTHENTICATION_FAILED, FRAUD_SUSPECTED, SUBSCRIPTION_FAILED, MANDATE_FAILED, CHECKOUT_ABANDONED) "
                f"and recovery suggestion (RETRY_SAME_METHOD, RETRY_DIFFERENT_METHOD, SEND_PAYMENT_LINK, WAIT_AND_RETRY, ESCALATE_TO_HUMAN, RETRY_SUBSCRIPTION, UPDATE_CARD_LINK, RENEW_MANDATE). "
                f"Return short diagnosis."
            )
            response = str(llm_agent(prompt))
            return {
                "category": "BANK_DECLINE",
                "suggestion": "SEND_PAYMENT_LINK",
                "root_cause": f"Strands AI Agent Diagnosis (mistral.ministral-3-8b-instruct): {response[:200]}...",
                "confidence": 0.90,
                "used_ai": True
            }
        except Exception as err:
            logging.warning(f"LLM diagnosis fallback error: {err}")

    return {
        "category": "UNKNOWN",
        "suggestion": "ESCALATE_TO_HUMAN",
        "root_cause": f"Ambiguous failure code ({error_code}): {description}. Flagged for AI/human diagnosis.",
        "confidence": 0.50,
        "used_ai": True
    }
