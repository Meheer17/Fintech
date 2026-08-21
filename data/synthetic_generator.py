import json
import random
import time

FAILURE_CATEGORIES = [
    ("BAD_REQUEST_PAYMENT_FAILED", "INSUFFICIENT_FUNDS", 0.35),
    ("CARD_EXPIRED", "CARD_EXPIRED", 0.20),
    ("BANK_TECHNICAL_GLITCH", "BANK_DECLINE", 0.20),
    ("AUTHENTICATION_FAILED", "AUTHENTICATION_FAILED", 0.15),
    ("FRAUD_SUSPECTED", "FRAUD_SUSPECTED", 0.10),
]

PAYMENT_METHODS = ["upi", "card", "netbanking", "wallet"]

def generate_synthetic_data(count: int = 100) -> dict:
    orders = []
    payments = []
    settlements = []

    for i in range(1, count + 1):
        order_id = f"order_RZP_{i:04d}"
        payment_id = f"pay_RZP_{i:04d}"
        settlement_id = f"set_RZP_{i:04d}"
        amount_paise = random.choice([50000, 120000, 250000, 499000, 1000000]) # ₹500 - ₹10,000
        method = random.choice(PAYMENT_METHODS)
        status = random.choices(["captured", "failed"], weights=[0.6, 0.4])[0]

        order = {
            "order_id": order_id,
            "amount_paise": amount_paise,
            "currency": "INR",
            "created_at": time.strftime("%Y-%m-%dT%H:%M:%SZ")
        }
        orders.append(order)

        err_code, category, _ = random.choice(FAILURE_CATEGORIES) if status == "failed" else ("", "", 0)
        payment = {
            "payment_id": payment_id,
            "order_id": order_id,
            "amount_paise": amount_paise,
            "method": method,
            "status": status,
            "error_code": err_code,
            "error_description": f"Razorpay payment {status}: {err_code}" if status == "failed" else "",
            "category": category
        }
        payments.append(payment)

        # Settlement (exact or fee-deducted)
        settled_amt = int(amount_paise * 0.98) if status == "captured" else 0
        settlement = {
            "settlement_id": settlement_id,
            "order_id": order_id,
            "payment_id": payment_id,
            "amount_paise": settled_amt,
            "utr": f"UTR_{random.randint(10000000, 99999999)}"
        }
        settlements.append(settlement)

    return {
        "orders": orders,
        "payments": payments,
        "settlements": settlements
    }

if __name__ == "__main__":
    data = generate_synthetic_data(100)
    with open("data/synthetic_dataset.json", "w") as f:
        json.dump(data, f, indent=2)
    print(f"Generated 100 synthetic Razorpay dataset records in data/synthetic_dataset.json")
