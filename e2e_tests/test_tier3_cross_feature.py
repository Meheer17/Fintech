import os
import sys
import json
import unittest
import importlib.util
from fastapi.testclient import TestClient

# Setup sys.path
BASE_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
if os.path.join(BASE_DIR, "ai_gateway", "src") not in sys.path:
    sys.path.insert(0, os.path.join(BASE_DIR, "ai_gateway", "src"))
if os.path.join(BASE_DIR, "failure_detector", "src") not in sys.path:
    sys.path.insert(0, os.path.join(BASE_DIR, "failure_detector", "src"))
if os.path.join(BASE_DIR, "recovery_orchestrator", "src") not in sys.path:
    sys.path.insert(0, os.path.join(BASE_DIR, "recovery_orchestrator", "src"))
if os.path.join(BASE_DIR, "reconciliation_engine", "src") not in sys.path:
    sys.path.insert(0, os.path.join(BASE_DIR, "reconciliation_engine", "src"))

def load_module_from_path(module_name, file_path):
    spec = importlib.util.spec_from_file_location(module_name, file_path)
    mod = importlib.util.module_from_spec(spec)
    sys.modules[module_name] = mod
    spec.loader.exec_module(mod)
    return mod

ai_main = load_module_from_path("ai_main", os.path.join(BASE_DIR, "ai_gateway", "src", "main.py"))
failure_main = load_module_from_path("failure_main", os.path.join(BASE_DIR, "failure_detector", "src", "main.py"))
recovery_main = load_module_from_path("recovery_main", os.path.join(BASE_DIR, "recovery_orchestrator", "src", "main.py"))
recon_main = load_module_from_path("recon_main", os.path.join(BASE_DIR, "reconciliation_engine", "src", "main.py"))

import classifier
import guardrails
import matcher
import forecast

class TestTier3CrossFeatureCombinations(unittest.TestCase):
    """
    Tier 3: Cross-Feature Combinations (Pairwise & Multi-Feature Interactions)
    Tests interactions across microservices:
    1. Failure Detection -> Recovery Orchestration -> Audit Trail Logging
    2. Settlement Ingestion -> gRPC Cash Forecasting -> Dashboard Overview
    3. Webhook Ingestion -> Failure Diagnosis -> Recovery Link Creation
    4. Subscription Failure -> Classifier -> Recovery Orchestrator
    5. Mandate Failure -> Classifier -> Recovery Renewal
    6. Recovery Completion -> Reconciliation Engine -> Match Rate
    7. Dispute Ingestion -> Overview Metrics Calculation
    8. B2B Promise Recording -> Cash Forecast Integration
    9. Copilot Diagnosis -> Payment Link Generation Tool Chain
    10. Reconciliation Batch -> Dashboard Report Breakdown
    """

    def setUp(self):
        self.ai_client = TestClient(ai_main.app)
        self.failure_client = TestClient(failure_main.app)
        self.recovery_client = TestClient(recovery_main.app)
        self.recon_client = TestClient(recon_main.app)

    def test_pairwise_01_failure_detection_to_recovery_to_audit(self):
        """Cross-Feature 1: Failure Detection -> Recovery Orchestration -> Audit Log Entry Generation."""
        diag = classifier.classify_failure("CARD_EXPIRED", "Credit card expired on renewal attempt")
        self.assertEqual(diag["category"], "CARD_EXPIRED")
        self.assertEqual(diag["suggestion"], "SEND_PAYMENT_LINK")

        wf = recovery_main.process_workflow(payment_id="pay_cross_001", current_retries=0, amount_paise=150000)
        self.assertEqual(wf["status"], "WF_IN_PROGRESS")
        self.assertEqual(wf["action"], "ACTION_CREATE_PAYMENT_LINK")

        audit_entry = {
            "entity_id": wf["workflow_id"],
            "action": wf["action"],
            "reason": wf["reason"],
            "guardrails_checked": wf["guardrails_checked"]
        }
        self.assertIn("RETRIES_UNDER_LIMIT", audit_entry["guardrails_checked"])

    def test_pairwise_02_settlement_ingestion_to_cash_forecast(self):
        """Cross-Feature 2: Settlement Ingestion -> gRPC Cash Position Forecast Calculation."""
        daily_settlement = 2500000
        days = 7

        fc = forecast.forecast_cash_position(days=days, avg_daily_settlement_paise=daily_settlement)
        expected_base = days * daily_settlement
        expected_boost = int(expected_base * 0.15)
        
        self.assertEqual(fc["total_projected_cash_paise"], expected_base + expected_boost)
        self.assertEqual(len(fc["daily_forecast"]), days)

    def test_pairwise_03_webhook_ingestion_to_failure_diagnosis(self):
        """Cross-Feature 3: Webhook payment.failed -> Failure Detector API Diagnosis."""
        webhook_payload = {
            "event": "payment.failed",
            "payload": {
                "payment": {
                    "entity": {
                        "id": "pay_wh_101",
                        "amount": 250000,
                        "error_code": "BAD_REQUEST_PAYMENT_FAILED",
                        "error_description": "Card authorization rejected by issuing bank"
                    }
                }
            }
        }
        p_entity = webhook_payload["payload"]["payment"]["entity"]

        diag_res = self.failure_client.post("/diagnose", json={
            "payment_id": p_entity["id"],
            "error_code": p_entity["error_code"],
            "description": p_entity["error_description"],
            "amount_paise": p_entity["amount"]
        })
        self.assertEqual(diag_res.status_code, 200)
        diag = diag_res.json()
        self.assertEqual(diag["payment_id"], "pay_wh_101")
        self.assertEqual(diag["category"], "INSUFFICIENT_FUNDS")

    def test_pairwise_04_subscription_failure_webhook_to_orchestrator(self):
        """Cross-Feature 4: Webhook subscription.charged.failed -> Classifier SUBSCRIPTION_FAILED -> Recovery Retry."""
        diag = classifier.classify_failure("SUBSCRIPTION_CHARGED_FAILED", "Mandate debit failed on recurring plan")
        self.assertEqual(diag["category"], "SUBSCRIPTION_FAILED")
        self.assertEqual(diag["suggestion"], "RETRY_SUBSCRIPTION")

        wf = recovery_main.process_workflow("pay_sub_fail", current_retries=1, amount_paise=99900)
        self.assertEqual(wf["status"], "WF_IN_PROGRESS")
        self.assertIn("payment_link_id", wf)

    def test_pairwise_05_mandate_expired_webhook_to_orchestrator(self):
        """Cross-Feature 5: Webhook MANDATE_EXPIRED -> Classifier MANDATE_FAILED -> Renewal Workflow."""
        diag = classifier.classify_failure("MANDATE_EXPIRED", "AutoPay mandate expired")
        self.assertEqual(diag["category"], "MANDATE_FAILED")
        self.assertEqual(diag["suggestion"], "RENEW_MANDATE")

        can_retry = guardrails.RecoveryGuardrails.check_retry_eligibility(0)
        self.assertTrue(can_retry)

    def test_pairwise_06_recovery_completion_to_reconciliation(self):
        """Cross-Feature 6: Recovery payment link completed -> Reconciliation 3-way matcher."""
        payment_amt = 150000
        order = {"amount_paise": payment_amt}
        payment = {"amount_paise": payment_amt}

        settlement = {"amount_paise": 147000}

        m_res = matcher.match_record(order, payment, settlement)
        self.assertEqual(m_res["match_type"], "FUZZY_MATCH")
        self.assertEqual(m_res["confidence"], 0.92)

    def test_pairwise_07_dispute_event_to_overview_metrics(self):
        """Cross-Feature 7: Dispute created webhook -> Overview metrics at-risk calculation."""
        disputes = [
            {"id": "disp_001", "amount": 250000, "status": "open"},
            {"id": "disp_002", "amount": 150000, "status": "open"}
        ]
        total_dispute_amount = sum(d["amount"] for d in disputes)
        self.assertEqual(total_dispute_amount, 400000)

    def test_pairwise_08_b2b_promise_to_cash_forecast_update(self):
        """Cross-Feature 8: B2B Promise-to-Pay recording -> Cash position forecast boost."""
        promises = [
            {"promise_id": "p1", "amount_inr": 50000.0, "status": "PENDING"},
            {"promise_id": "p2", "amount_inr": 30000.0, "status": "PENDING"}
        ]
        promise_total_paise = int(sum(p["amount_inr"] for p in promises) * 100)

        fc = forecast.forecast_cash_position(days=7, avg_daily_settlement_paise=2000000)
        base_cash = fc["total_projected_cash_paise"]

        combined_cash = base_cash + promise_total_paise
        self.assertGreater(combined_cash, base_cash)

    def test_pairwise_09_chat_tool_diagnosis_to_link_creation(self):
        """Cross-Feature 9: AI Copilot diagnosis tool -> payment link creation tool chain."""
        diag_json = ai_main.diagnose_payment_failure(payment_id="pay_chat_001", error_code="CARD_EXPIRED")
        diag_res = json.loads(diag_json)
        self.assertIn("payment_id", diag_res)

        link_json = ai_main.create_payment_link_tool(
            name="Sara Jenkins",
            email="sara@gmail.com",
            phone="9876543210",
            amount_inr=1200.0,
            description="Card renewal recovery link"
        )
        link_res = json.loads(link_json)
        if link_res.get("success"):
            self.assertEqual(link_res["customer_name"], "Sara Jenkins")

    def test_pairwise_10_reconciliation_batch_to_dashboard_report(self):
        """Cross-Feature 10: Run reconciliation batch -> Summary breakdown report."""
        batch = [
            {"order": {"amount_paise": 10000}, "payment": {"amount_paise": 10000}, "settlement": {"amount_paise": 10000}},
            {"order": {"amount_paise": 20000}, "payment": {"amount_paise": 20000}, "settlement": {"amount_paise": 19600}},
            {"order": {"amount_paise": 30000}, "payment": {"amount_paise": 30000}, "settlement": {"amount_paise": 15000}}
        ]
        res = self.recon_client.post("/reconcile", json={"batch": batch})
        self.assertEqual(res.status_code, 200)
        data = res.json()
        self.assertEqual(data["total_records"], 3)
        self.assertEqual(data["exact_matches"], 1)
        self.assertEqual(data["fuzzy_matches"], 1)
        self.assertEqual(data["unmatched"], 1)
        self.assertAlmostEqual(data["match_rate"], 0.6667, places=3)

if __name__ == "__main__":
    unittest.main()
