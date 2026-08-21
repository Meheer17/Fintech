# RevenueIQ Tiered Three-Way Settlement Matcher

def match_record(order: dict, payment: dict, settlement: dict) -> dict:
    order_amt = order.get("amount_paise", 0)
    payment_amt = payment.get("amount_paise", 0)
    settled_amt = settlement.get("amount_paise", 0)

    # 1. Exact Match
    if order_amt == payment_amt and payment_amt == settled_amt:
        return {
            "match_type": "EXACT_MATCH",
            "confidence": 1.0,
            "notes": "Exact match across Order, Payment, and Settlement."
        }
    
    # 2. Fuzzy Match (within ±2% card processing fee tolerance)
    expected_after_fee = int(payment_amt * 0.98)
    if abs(settled_amt - expected_after_fee) <= 500: # ±₹5.00
        return {
            "match_type": "FUZZY_MATCH",
            "confidence": 0.92,
            "notes": f"Fuzzy match with fee tolerance: expected ₹{expected_after_fee/100:.2f}, settled ₹{settled_amt/100:.2f}"
        }

    # 3. AI Exception Flag
    return {
        "match_type": "UNMATCHED",
        "confidence": 0.0,
        "notes": f"Discrepancy: Order=₹{order_amt/100:.2f}, Payment=₹{payment_amt/100:.2f}, Settled=₹{settled_amt/100:.2f}"
    }
