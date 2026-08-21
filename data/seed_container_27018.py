import json
import pymongo
from datetime import datetime, timedelta

def seed_mongo_27018():
    client = pymongo.MongoClient("mongodb://localhost:27018")
    print("Connecting to MongoDB container via host port 27018...")

    with open("data/synthetic_dataset.json", "r") as f:
        dataset = json.load(f)

    orders = dataset.get("orders", [])
    payments = dataset.get("payments", [])
    settlements = dataset.get("settlements", [])

    for db_name in ["revenueiq_db", "mongodb_service_db"]:
        db = client[db_name]
        
        db["orders"].delete_many({})
        if orders:
            db["orders"].insert_many(orders)
        
        db["payments"].delete_many({})
        if payments:
            db["payments"].insert_many(payments)

        db["settlements"].delete_many({})
        if settlements:
            db["settlements"].insert_many(settlements)

        db["failures"].delete_many({})
        failed_payments = [p for p in payments if p.get("status") == "failed"]
        failures = []
        now = datetime.utcnow()
        for idx, fp in enumerate(failed_payments):
            status = "RECOVERED" if idx % 3 == 0 else ("IN_PROGRESS" if idx % 3 == 1 else "PENDING")
            failures.append({
                "id": f"fail_{100 + idx}",
                "payment_id": fp["payment_id"],
                "order_id": fp["order_id"],
                "amount_paise": fp["amount_paise"],
                "payment_method": fp.get("method", "upi"),
                "category": fp.get("category", "BANK_DECLINE"),
                "error_code": fp.get("error_code", "BAD_REQUEST"),
                "error_description": fp.get("error_description", ""),
                "root_cause": f"Systemic {fp.get('category', 'BANK_DECLINE')} detected on {fp.get('method', 'upi').upper()} rail",
                "suggestion": "RETRY_PAYMENT" if fp.get("method") == "upi" else "SEND_PAYMENT_LINK",
                "recovery_status": status,
                "failed_at": (now - timedelta(hours=idx * 2)).isoformat() + "Z"
            })
        if failures:
            db["failures"].insert_many(failures)

        db["workflows"].delete_many({})
        workflows = []
        for idx, f in enumerate(failures[:15]):
            wf_status = "WF_RECOVERED" if f["recovery_status"] == "RECOVERED" else ("WF_IN_PROGRESS" if f["recovery_status"] == "IN_PROGRESS" else "WF_CREATED")
            workflows.append({
                "workflow_id": f"wf_{800 + idx}",
                "payment_id": f["payment_id"],
                "amount_paise": f["amount_paise"],
                "workflow_status": wf_status,
                "retry_count": 1 if wf_status != "WF_CREATED" else 0,
                "recovered_id": f"pay_REC_{idx:03d}" if wf_status == "WF_RECOVERED" else "",
                "guardrails_checked": ["MAX_RETRIES", "CONTACT_WINDOW", "COST_CAP"],
                "steps": [
                    {"step_id": f"st_{idx}_1", "action": "DIAGNOSE_FAILURE", "status": "SUCCESS", "result": f"Category: {f['category']} (95% confidence)", "executed_at": f["failed_at"]},
                    {"step_id": f"st_{idx}_2", "action": f["suggestion"], "status": "SUCCESS" if wf_status == "WF_RECOVERED" else "PENDING", "result": f"Action executed: {f['suggestion']}", "executed_at": (now - timedelta(hours=idx)).isoformat() + "Z"}
                ],
                "created_at": f["failed_at"]
            })
        if workflows:
            db["workflows"].insert_many(workflows)

        db["audit_logs"].delete_many({})
        audit_logs = []
        for idx, f in enumerate(failures[:20]):
            audit_logs.append({
                "id": f"aud_{200 + idx}",
                "timestamp": f["failed_at"],
                "service_name": "failure-detector",
                "action": "DIAGNOSE_FAILURE",
                "entity_type": "PAYMENT",
                "entity_id": f["payment_id"],
                "status": "COMPLETED",
                "reasoning": f["root_cause"],
                "actor": "diagnosis_agent",
                "guardrails_checked": ["HMAC_VERIFICATION", "CONFIDENCE_THRESHOLD"]
            })
        if audit_logs:
            db["audit_logs"].insert_many(audit_logs)

        db["promises"].delete_many({})
        db["promises"].insert_many([
            {"id": "prm_301", "customer": "Vikram Mehta", "amount_paise": 350000, "promised_date": "2026-08-25", "status": "KEPT", "workflow_id": "wf_801"},
            {"id": "prm_302", "customer": "Ananya Roy", "amount_paise": 890000, "promised_date": "2026-08-28", "status": "PENDING", "workflow_id": "wf_802"},
            {"id": "prm_303", "customer": "Rahul Deshmukh", "amount_paise": 1250000, "promised_date": "2026-08-30", "status": "PENDING", "workflow_id": "wf_803"}
        ])

    print("Successfully populated MongoDB databases (revenueiq_db & mongodb_service_db) on port 27018!")

if __name__ == "__main__":
    seed_mongo_27018()
