import json
import pymongo
import time
from datetime import datetime, timedelta

def seed_mongodb():
    client = pymongo.MongoClient("mongodb://localhost:27017")
    db = client["revenueiq_db"]
    
    # Load synthetic dataset
    with open("data/synthetic_dataset.json", "r") as f:
        dataset = json.load(f)

    orders = dataset.get("orders", [])
    payments = dataset.get("payments", [])
    settlements = dataset.get("settlements", [])

    # Seed Orders
    db["orders"].delete_many({})
    if orders:
        db["orders"].insert_many(orders)
    print(f"Seeded {len(orders)} orders into MongoDB.")

    # Seed Payments
    db["payments"].delete_many({})
    if payments:
        db["payments"].insert_many(payments)
    print(f"Seeded {len(payments)} payments into MongoDB.")

    # Seed Settlements
    db["settlements"].delete_many({})
    if settlements:
        db["settlements"].insert_many(settlements)
    print(f"Seeded {len(settlements)} settlements into MongoDB.")

    # Seed Failures from failed payments
    db["failures"].delete_many({})
    failed_payments = [p for p in payments if p.get("status") == "failed"]
    failures = []
    now = datetime.utcnow()
    
    for idx, fp in enumerate(failed_payments):
        status = "RECOVERED" if idx % 3 == 0 else ("IN_PROGRESS" if idx % 3 == 1 else "PENDING")
        failure = {
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
        }
        failures.append(failure)

    if failures:
        db["failures"].insert_many(failures)
    print(f"Seeded {len(failures)} failure records into MongoDB.")

    # Seed Workflows
    db["workflows"].delete_many({})
    workflows = []
    for idx, f in enumerate(failures[:15]):
        wf_status = "WF_RECOVERED" if f["recovery_status"] == "RECOVERED" else ("WF_IN_PROGRESS" if f["recovery_status"] == "IN_PROGRESS" else "WF_CREATED")
        wf = {
            "workflow_id": f"wf_{800 + idx}",
            "payment_id": f["payment_id"],
            "amount_paise": f["amount_paise"],
            "workflow_status": wf_status,
            "retry_count": 1 if wf_status != "WF_CREATED" else 0,
            "recovered_id": f"pay_REC_{idx:03d}" if wf_status == "WF_RECOVERED" else "",
            "guardrails_checked": ["MAX_RETRIES", "CONTACT_WINDOW", "COST_CAP"],
            "steps": [
                {
                    "step_id": f"st_{idx}_1",
                    "action": "DIAGNOSE_FAILURE",
                    "status": "SUCCESS",
                    "result": f"Category: {f['category']} (95% confidence)",
                    "executed_at": f["failed_at"]
                },
                {
                    "step_id": f"st_{idx}_2",
                    "action": f["suggestion"],
                    "status": "SUCCESS" if wf_status == "WF_RECOVERED" else "PENDING",
                    "result": f"Action executed: {f['suggestion']}",
                    "executed_at": (now - timedelta(hours=idx)).isoformat() + "Z"
                }
            ],
            "created_at": f["failed_at"]
        }
        workflows.append(wf)

    if workflows:
        db["workflows"].insert_many(workflows)
    print(f"Seeded {len(workflows)} recovery workflows into MongoDB.")

    # Seed Audit Trail
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
        audit_logs.append({
            "id": f"aud_wf_{200 + idx}",
            "timestamp": (now - timedelta(hours=idx)).isoformat() + "Z",
            "service_name": "recovery-orchestrator",
            "action": f["suggestion"],
            "entity_type": "WORKFLOW",
            "entity_id": f["payment_id"],
            "status": "SUCCESS" if f["recovery_status"] == "RECOVERED" else "IN_PROGRESS",
            "reasoning": f"Executed {f['suggestion']} under guardrails max_retries=3",
            "actor": "strategy_agent",
            "guardrails_checked": ["MAX_RETRIES", "CONTACT_WINDOW", "COST_CAP"]
        })

    if audit_logs:
        db["audit_logs"].insert_many(audit_logs)
    print(f"Seeded {len(audit_logs)} audit trail entries into MongoDB.")

    # Seed Promises
    db["promises"].delete_many({})
    promises = [
        {
            "id": "prm_301",
            "customer": "Vikram Mehta",
            "amount_paise": 350000,
            "promised_date": "2026-08-25",
            "status": "KEPT",
            "workflow_id": "wf_801"
        },
        {
            "id": "prm_302",
            "customer": "Ananya Roy",
            "amount_paise": 890000,
            "promised_date": "2026-08-28",
            "status": "PENDING",
            "workflow_id": "wf_802"
        },
        {
            "id": "prm_303",
            "customer": "Rahul Deshmukh",
            "amount_paise": 1250000,
            "promised_date": "2026-08-30",
            "status": "PENDING",
            "workflow_id": "wf_803"
        }
    ]
    db["promises"].insert_many(promises)
    print(f"Seeded {len(promises)} promise records into MongoDB.")

    # Seed Disputes
    db["disputes"].delete_many({})
    disputes = [
        {"id": "disp_RZP_001", "dispute_id": "disp_RZP_001", "payment_id": "pay_RZP_0012", "amount": 250000, "amount_paise": 250000, "currency": "INR", "reason_code": "FRAUDULENT", "reason": "Customer claims payment was unauthorized", "phase": "chargeback", "status": "under_review", "respond_by": int((now + timedelta(days=3)).timestamp()), "created_at": int((now - timedelta(days=2)).timestamp())},
        {"id": "disp_RZP_002", "dispute_id": "disp_RZP_002", "payment_id": "pay_RZP_0024", "amount": 499000, "amount_paise": 499000, "currency": "INR", "reason_code": "SERVICES_NOT_PROVIDED", "reason": "Services promised were not delivered on time", "phase": "pre_arbitration", "status": "needs_response", "respond_by": int((now + timedelta(days=5)).timestamp()), "created_at": int((now - timedelta(days=1)).timestamp())},
        {"id": "disp_RZP_003", "dispute_id": "disp_RZP_003", "payment_id": "pay_RZP_0036", "amount": 120000, "amount_paise": 120000, "currency": "INR", "reason_code": "DUPLICATE_CHARGE", "reason": "Billed twice for single transaction", "phase": "retrieval", "status": "won", "respond_by": int((now - timedelta(days=4)).timestamp()), "created_at": int((now - timedelta(days=7)).timestamp())},
        {"id": "disp_RZP_004", "dispute_id": "disp_RZP_004", "payment_id": "pay_RZP_0048", "amount": 1000000, "amount_paise": 1000000, "currency": "INR", "reason_code": "CREDIT_NOT_PROCESSED", "reason": "Merchant agreed refund not processed", "phase": "chargeback", "status": "lost", "respond_by": int((now - timedelta(days=10)).timestamp()), "created_at": int((now - timedelta(days=14)).timestamp())},
        {"id": "disp_RZP_005", "dispute_id": "disp_RZP_005", "payment_id": "pay_RZP_0060", "amount": 350000, "amount_paise": 350000, "currency": "INR", "reason_code": "UNRECOGNIZED", "reason": "Transaction name unrecognized on statement", "phase": "retrieval", "status": "open", "respond_by": int((now + timedelta(days=6)).timestamp()), "created_at": int((now - timedelta(hours=12)).timestamp())}
    ]
    db["disputes"].insert_many(disputes)
    print(f"Seeded {len(disputes)} dispute records into MongoDB.")

    # Also seed mongodb_service_db database so mongo_service container can read it as well
    db2 = client["mongodb_service_db"]
    for col in ["orders", "payments", "settlements", "failures", "workflows", "audit_logs", "promises", "disputes"]:
        db2[col].delete_many({})
        docs = list(db[col].find({}, {"_id": 0}))
        if docs:
            db2[col].insert_many(docs)
    print("Copied all collections into mongodb_service_db.")

if __name__ == "__main__":
    seed_mongodb()
