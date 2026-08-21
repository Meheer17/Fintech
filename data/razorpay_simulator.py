import hmac
import hashlib
import json
import time
import urllib.request
import urllib.parse
import sys

WEBHOOK_URL = "http://localhost:8001/webhook/razorpay"
WEBHOOK_SECRET = "rzp_webhook_secret_123"

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
			print(f"[SIMULATOR] Sent {event_type} | Response: {res_body}")
			return True
	except Exception as e:
		print(f"[SIMULATOR WARNING] Could not connect to webhook-receiver ({e}). Is container running?")
		return False

def simulate_batch(json_file: str = "data/synthetic_dataset.json"):
	try:
		with open(json_file, "r") as f:
			dataset = json.load(f)
	except Exception as e:
		print(f"[SIMULATOR ERROR] Failed to load dataset: {e}")
		return

	payments = dataset.get("payments", [])
	print(f"Simulating {len(payments)} Razorpay webhooks...")

	for p in payments:
		event_type = "payment.failed" if p.get("status") == "failed" else "payment.captured"
		event_payload = {
			"event": event_type,
			"account_id": f"acc_merchant_{p.get('order_id', '1')}",
			"contains": ["payment"],
			"payload": {
				"payment": {
					"entity": {
						"id": p.get("payment_id"),
						"order_id": p.get("order_id"),
						"amount": p.get("amount_paise"),
						"currency": "INR",
						"status": p.get("status"),
						"method": p.get("method"),
						"error_code": p.get("error_code"),
						"error_description": p.get("error_description")
					}
				}
			}
		}
		send_webhook(event_type, event_payload)
		time.sleep(0.1)

if __name__ == "__main__":
	simulate_batch()
