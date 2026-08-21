# RevenueIQ Failure Classifier (Rule Engine + AI Fallback)

ERROR_CODE_MAP = {
    "BAD_REQUEST_PAYMENT_FAILED": ("INSUFFICIENT_FUNDS", "RETRY_SAME_METHOD"),
    "GATEWAY_ERROR": ("NETWORK_ERROR", "WAIT_AND_RETRY"),
    "CARD_EXPIRED": ("CARD_EXPIRED", "SEND_PAYMENT_LINK"),
    "AUTHENTICATION_FAILED": ("AUTHENTICATION_FAILED", "RETRY_DIFFERENT_METHOD"),
    "BANK_TECHNICAL_GLITCH": ("BANK_DECLINE", "WAIT_AND_RETRY"),
    "FRAUD_SUSPECTED": ("FRAUD_SUSPECTED", "ESCALATE_TO_HUMAN"),
    "LIMIT_EXCEEDED": ("LIMIT_EXCEEDED", "CONTACT_CUSTOMER"),
    "SUBSCRIPTION_CHARGED_FAILED": ("SUBSCRIPTION_FAILED", "RETRY_SUBSCRIPTION"),
    "MANDATE_EXPIRED": ("MANDATE_FAILED", "RENEW_MANDATE"),
    "CHECKOUT_ABANDONED": ("CHECKOUT_ABANDONED", "CHECKOUT_NUDGE"),
}

def classify_failure(error_code: str, description: str = "") -> dict:
    code_upper = error_code.upper() if error_code else "UNKNOWN"
    for key, (category, suggestion) in ERROR_CODE_MAP.items():
        if key in code_upper:
            return {
                "category": category,
                "suggestion": suggestion,
                "root_cause": f"Diagnosed by rule engine: {description or category}",
                "confidence": 0.95,
                "used_ai": False
            }
    
    # Fallback diagnosis
    return {
        "category": "UNKNOWN",
        "suggestion": "ESCALATE_TO_HUMAN",
        "root_cause": f"Ambiguous failure code ({error_code}): {description}. Flagged for AI/human diagnosis.",
        "confidence": 0.50,
        "used_ai": True
    }
