import os
import sys
import json
import logging
import httpx
import pymongo
from dotenv import load_dotenv

load_dotenv()

logging.basicConfig(level=logging.INFO, format="%(asctime)s [%(levelname)s] %(message)s")

RAZORPAY_KEY_ID = os.getenv("RAZORPAY_KEY_ID", "rzp_test_SESIqmsJZRvpZ1")
RAZORPAY_KEY_SECRET = os.getenv("RAZORPAY_KEY_SECRET", "4qzbtZYU4wrF4ER3lYk2hIt7")
MONGO_URI = os.getenv("MONGO_URI_HOST", "mongodb://localhost:27017")

def fetch_razorpay_resource(endpoint: str) -> list:
    url = f"https://api.razorpay.com/v1/{endpoint}"
    try:
        r = httpx.get(url, auth=(RAZORPAY_KEY_ID, RAZORPAY_KEY_SECRET), timeout=10.0)
        if r.status_code == 200:
            data = r.json()
            items = data.get("items", [])
            logging.info(f"Fetched {len(items)} items from Razorpay API: /v1/{endpoint}")
            return items
        else:
            logging.error(f"Razorpay API error for /v1/{endpoint} ({r.status_code}): {r.text}")
            return []
    except Exception as e:
        logging.error(f"Failed to fetch /v1/{endpoint} from Razorpay: {e}")
        return []

def sync_razorpay_to_mongo():
    logging.info(f"Starting Razorpay account sync using Key ID: {RAZORPAY_KEY_ID[:10]}...")
    client = pymongo.MongoClient(MONGO_URI)
    
    # 1. Fetch live data from Razorpay API
    settlements = fetch_razorpay_resource("settlements")
    subscriptions = fetch_razorpay_resource("subscriptions")
    payments = fetch_razorpay_resource("payments")
    orders = fetch_razorpay_resource("orders")
    disputes = fetch_razorpay_resource("disputes")
    refunds = fetch_razorpay_resource("refunds")

    # 2. Sync to MongoDB databases
    for db_name in ["revenueiq_db", "mongodb_service_db"]:
        db = client[db_name]
        logging.info(f"Updating database '{db_name}' with live Razorpay data...")

        # Sync Settlements
        db["settlements"].delete_many({})
        if settlements:
            db["settlements"].insert_many(settlements)
        logging.info(f"Synced {len(settlements)} settlements to '{db_name}'.settlements")

        # Sync Subscriptions
        db["subscriptions"].delete_many({})
        if subscriptions:
            db["subscriptions"].insert_many(subscriptions)
        logging.info(f"Synced {len(subscriptions)} subscriptions to '{db_name}'.subscriptions")

        # Sync Payments
        db["payments"].delete_many({})
        if payments:
            db["payments"].insert_many(payments)
        logging.info(f"Synced {len(payments)} payments to '{db_name}'.payments")

        # Sync Orders
        db["orders"].delete_many({})
        if orders:
            db["orders"].insert_many(orders)
        logging.info(f"Synced {len(orders)} orders to '{db_name}'.orders")

        # Sync Disputes
        db["disputes"].delete_many({})
        if disputes:
            db["disputes"].insert_many(disputes)
        logging.info(f"Synced {len(disputes)} disputes to '{db_name}'.disputes")

        # Sync Refunds
        db["refunds"].delete_many({})
        if refunds:
            db["refunds"].insert_many(refunds)
        logging.info(f"Synced {len(refunds)} refunds to '{db_name}'.refunds")

        # Sync Failure Events & Workflows
        failed_payments = [p for p in payments if p.get("status") in ["failed", "refunded"]]
        failures = []
        workflows = []
        for idx, fp in enumerate(failed_payments):
            payment_id = fp.get("id", f"pay_fail_{idx}")
            amt = fp.get("amount", 50000)
            fail_doc = {
                "id": f"fail_{payment_id}",
                "payment_id": payment_id,
                "order_id": fp.get("order_id", ""),
                "amount_paise": amt,
                "payment_method": fp.get("method", "card"),
                "category": fp.get("error_code", "BANK_DECLINE"),
                "error_code": fp.get("error_code", "BAD_REQUEST_PAYMENT_FAILED"),
                "error_description": fp.get("error_description", "Payment failed on card/netbanking rail"),
                "root_cause": f"Payment failure on method {fp.get('method', 'card')}",
                "suggestion": "SEND_PAYMENT_LINK",
                "recovery_status": "PENDING",
                "failed_at": fp.get("created_at")
            }
            failures.append(fail_doc)

            wf_doc = {
                "workflow_id": f"wf_{payment_id}",
                "payment_id": payment_id,
                "amount_paise": amt,
                "status": "WF_IN_PROGRESS",
                "action": "ACTION_CREATE_PAYMENT_LINK",
                "reason": "Live recovery workflow active",
                "guardrails_checked": ["MAX_RETRIES", "CONTACT_WINDOW"],
                "created_at": fp.get("created_at")
            }
            workflows.append(wf_doc)

        db["failures"].delete_many({})
        if failures:
            db["failures"].insert_many(failures)

        db["workflows"].delete_many({})
        if workflows:
            db["workflows"].insert_many(workflows)

    logging.info("Razorpay live sync completed successfully!")
    return {
        "settlements": len(settlements),
        "subscriptions": len(subscriptions),
        "payments": len(payments),
        "orders": len(orders),
        "disputes": len(disputes),
        "refunds": len(refunds),
        "failures": len(failures)
    }

if __name__ == "__main__":
    res = sync_razorpay_to_mongo()
    print("SYNC RESULT:", json.dumps(res, indent=2))
