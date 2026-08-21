import hmac
import hashlib
import json
import time
import urllib.request
import os

WEBHOOK_URL = os.getenv("WEBHOOK_URL", "http://localhost:8001/webhook/razorpay")
WEBHOOK_SECRET = os.getenv("RAZORPAY_WEBHOOK_SECRET", "IwNdZ8/zcd7bXcE8v4v0BLXMCXj+d8lnXKU6s700Dhs=")

def send_webhook(event_type: str, payload: dict, secret: str = WEBHOOK_SECRET) -> bool:
    raw_body = json.dumps(payload).encode("utf-8")
    mac = hmac.new(secret.encode("utf-8"), raw_body, hashlib.sha256)
    signature = mac.hexdigest()

    req = urllib.request.Request(WEBHOOK_URL, data=raw_body, headers={
        "Content-Type": "application/json",
        "X-Razorpay-Signature": signature
    })

    try:
        with urllib.request.urlopen(req, timeout=5) as response:
            res_body = response.read().decode("utf-8")
            print(f"[SIMULATOR] Sent '{event_type}' -> Response: {res_body}")
            return True
    except Exception as e:
        print(f"[SIMULATOR WARNING] Could not send {event_type} to webhook-receiver ({e})")
        return False

def generate_all_10_events():
    timestamp = int(time.time())

    events = [
        # 1. payment.authorized
        ("payment.authorized", {
            "event": "payment.authorized",
            "account_id": "acc_merchant_001",
            "payload": {
                "payment": {
                    "entity": {
                        "id": f"pay_auth_{timestamp}",
                        "entity": "payment",
                        "amount": 250000,
                        "currency": "INR",
                        "status": "authorized",
                        "order_id": f"order_001_{timestamp}",
                        "method": "card",
                        "email": "customer1@example.com",
                        "contact": "+919876543210"
                    }
                }
            }
        }),

        # 2. payment.failed
        ("payment.failed", {
            "event": "payment.failed",
            "account_id": "acc_merchant_001",
            "payload": {
                "payment": {
                    "entity": {
                        "id": f"pay_fail_{timestamp}",
                        "entity": "payment",
                        "amount": 150000,
                        "currency": "INR",
                        "status": "failed",
                        "order_id": f"order_002_{timestamp}",
                        "method": "upi",
                        "error_code": "BAD_REQUEST_PAYMENT_TIMED_OUT",
                        "error_description": "Payment authorization timed out from UPI PSP bank",
                        "error_source": "bank",
                        "error_reason": "payment_verification_timeout"
                    }
                }
            }
        }),

        # 3. payment.captured
        ("payment.captured", {
            "event": "payment.captured",
            "account_id": "acc_merchant_001",
            "payload": {
                "payment": {
                    "entity": {
                        "id": f"pay_cap_{timestamp}",
                        "entity": "payment",
                        "amount": 250000,
                        "currency": "INR",
                        "status": "captured",
                        "order_id": f"order_001_{timestamp}",
                        "method": "card",
                        "fee": 5000,
                        "tax": 900
                    }
                }
            }
        }),

        # 4. payment.dispute.created
        ("payment.dispute.created", {
            "event": "payment.dispute.created",
            "account_id": "acc_merchant_001",
            "payload": {
                "dispute": {
                    "entity": {
                        "id": f"disp_{timestamp}",
                        "entity": "dispute",
                        "payment_id": f"pay_cap_{timestamp}",
                        "amount": 250000,
                        "currency": "INR",
                        "reason_code": "MERCHANT_NOT_PROVIDING_SERVICE",
                        "status": "open",
                        "respond_by": timestamp + 604800
                    }
                }
            }
        }),

        # 5. order.paid
        ("order.paid", {
            "event": "order.paid",
            "account_id": "acc_merchant_001",
            "payload": {
                "order": {
                    "entity": {
                        "id": f"order_001_{timestamp}",
                        "entity": "order",
                        "amount": 250000,
                        "amount_paid": 250000,
                        "status": "paid",
                        "attempts": 1
                    }
                }
            }
        }),

        # 6. subscription.pending
        ("subscription.pending", {
            "event": "subscription.pending",
            "account_id": "acc_merchant_001",
            "payload": {
                "subscription": {
                    "entity": {
                        "id": f"sub_pend_{timestamp}",
                        "entity": "subscription",
                        "plan_id": "plan_pro_monthly",
                        "customer_id": "cust_101",
                        "status": "pending",
                        "current_start": timestamp,
                        "current_end": timestamp + 2592000
                    }
                }
            }
        }),

        # 7. subscription.charged
        ("subscription.charged", {
            "event": "subscription.charged",
            "account_id": "acc_merchant_001",
            "payload": {
                "subscription": {
                    "entity": {
                        "id": f"sub_active_{timestamp}",
                        "entity": "subscription",
                        "plan_id": "plan_pro_monthly",
                        "customer_id": "cust_102",
                        "status": "active",
                        "paid_count": 3
                    }
                }
            }
        }),

        # 8. subscription.cancelled
        ("subscription.cancelled", {
            "event": "subscription.cancelled",
            "account_id": "acc_merchant_001",
            "payload": {
                "subscription": {
                    "entity": {
                        "id": f"sub_cancel_{timestamp}",
                        "entity": "subscription",
                        "plan_id": "plan_pro_monthly",
                        "customer_id": "cust_103",
                        "status": "cancelled",
                        "ended_at": timestamp
                    }
                }
            }
        }),

        # 9. settlement.processed
        ("settlement.processed", {
            "event": "settlement.processed",
            "account_id": "acc_merchant_001",
            "payload": {
                "settlement": {
                    "entity": {
                        "id": f"set_{timestamp}",
                        "entity": "settlement",
                        "amount": 244100,
                        "status": "processed",
                        "fees": 5000,
                        "tax": 900,
                        "utr": f"UTR_HDFC_{timestamp}"
                    }
                }
            }
        }),

        # 10. refund.created
        ("refund.created", {
            "event": "refund.created",
            "account_id": "acc_merchant_001",
            "payload": {
                "refund": {
                    "entity": {
                        "id": f"rfnd_{timestamp}",
                        "entity": "refund",
                        "payment_id": f"pay_cap_{timestamp}",
                        "amount": 50000,
                        "currency": "INR",
                        "status": "processed"
                    }
                }
            }
        })
    ]

    print(f"Simulating all 10 Active Razorpay Webhook Events to {WEBHOOK_URL}...")
    for ev_name, ev_payload in events:
        send_webhook(ev_name, ev_payload)
        time.sleep(0.1)

if __name__ == "__main__":
    generate_all_10_events()
