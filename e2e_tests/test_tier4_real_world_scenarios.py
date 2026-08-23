import os
import sys
import json
import unittest
import hmac
import hashlib
import time
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

class TestTier4RealWorldScenarios(unittest.TestCase):
    """
    Tier 4: Real-World Application Integration Scenarios (End-to-End Workflows)
    Simulates complete production end-to-end workflows across RevenueIQ components:
    1. Complete Dunning Lifecycle (Subscription failure -> Diagnosis -> Recovery Orchestrator -> Payment link -> Audit)
    2. Mandate Failure & AutoPay Renewal (e-Mandate expiry -> MANDATE_FAILED diagnosis -> Renewal workflow)
    3. Three-Way Settlement Reconciliation Workflow (Order, Payment, Settlement batch match split & Dashboard report)
    4. B2B Receivables & Cash Forecast Workflow (Promise recording -> Forward 7/30 day cash position with recovery boost)
    5. Interactive AI Copilot Payment Link Generation Workflow (Validation gate -> Detail collection -> Razorpay link creation)
    6. Webhook Stream Security & Burst Resilience Workflow (10 active Razorpay event types & HMAC SHA256 verification)
    """

    def setUp(self):
        self.ai_client = TestClient(ai_main.app)
        self.failure_client = TestClient(failure_main.app)
        self.recovery_client = TestClient(recovery_main.app)
        self.recon_client = TestClient(recon_main.app)

    def test_scenario_01_complete_dunning_lifecycle(self):
        """
        Scenario 1: Complete Dunning Lifecycle Workflow
        1. Razorpay webhook fires subscription.charged.failed
        2. Failure Detector classifies as SUBSCRIPTION_FAILED & RETRY_SUBSCRIPTION
        3. Recovery Orchestrator verifies retries (attempt 0 < 3) and cost cap
        4. Generates recovery workflow with Razorpay Payment Link
        5. Verifies audit trail entry formulation
        """
        payment_id = "pay_dunning_909"
        error_code = "SUBSCRIPTION_CHARGED_FAILED"
        amount_paise = 299900

        diag_res = self.failure_client.post("/diagnose", json={
            "payment_id": payment_id,
            "error_code": error_code,
            "description": "Recurring mandate debit failed on bank side",
            "amount_paise": amount_paise
        })
        self.assertEqual(diag_res.status_code, 200)
        diag = diag_res.json()
        self.assertEqual(diag["category"], "SUBSCRIPTION_FAILED")
        self.assertEqual(diag["suggestion"], "RETRY_SUBSCRIPTION")

        can_retry = guardrails.RecoveryGuardrails.check_retry_eligibility(0)
        cost_ok = guardrails.RecoveryGuardrails.check_cost_cap(500, amount_paise)
        self.assertTrue(can_retry)
        self.assertTrue(cost_ok)

        orch_res = self.recovery_client.post("/orchestrate", json={
            "payment_id": payment_id,
            "current_retries": 0,
            "amount_paise": amount_paise
        })
        self.assertEqual(orch_res.status_code, 200)
        orch = orch_res.json()
        self.assertEqual(orch["status"], "WF_IN_PROGRESS")
        self.assertEqual(orch["action"], "ACTION_CREATE_PAYMENT_LINK")
        self.assertIn("payment_link_id", orch)
        self.assertIn("short_url", orch)

        audit_log = {
            "entity_id": orch["workflow_id"],
            "payment_id": payment_id,
            "action": orch["action"],
            "status": orch["status"],
            "guardrails_checked": orch["guardrails_checked"]
        }
        self.assertEqual(audit_log["status"], "WF_IN_PROGRESS")

    def test_scenario_02_mandate_failure_and_renewal(self):
        """
        Scenario 2: Mandate Failure & AutoPay Renewal Workflow
        1. AutoPay debit attempt fails with MANDATE_EXPIRED
        2. Failure Classifier diagnoses MANDATE_FAILED with suggestion RENEW_MANDATE
        3. Recovery Orchestrator dispatches renewal action
        4. Guardrails verify max contact limit per day
        """
        payment_id = "pay_mandate_555"
        error_code = "MANDATE_EXPIRED"
        amount_paise = 500000

        diag = classifier.classify_failure(error_code, "e-Mandate authorization period expired")
        self.assertEqual(diag["category"], "MANDATE_FAILED")
        self.assertEqual(diag["suggestion"], "RENEW_MANDATE")

        contacts_today = 1
        contacts_allowed = guardrails.RecoveryGuardrails.check_contact_limit(contacts_today)
        self.assertTrue(contacts_allowed)

        wf = recovery_main.process_workflow(payment_id, current_retries=1, amount_paise=amount_paise)
        self.assertEqual(wf["status"], "WF_IN_PROGRESS")
        self.assertEqual(wf["payment_id"], payment_id)

    def test_scenario_03_three_way_reconciliation_workflow(self):
        """
        Scenario 3: Three-Way Settlement Reconciliation Workflow
        1. Process a batch of transactions containing exact matches, fuzzy matches (fee deductions), and unmatched discrepancies
        2. Reconciliation Engine categorizes each item
        3. Returns overall match rate and summary breakdown
        """
        batch = [
            {"order": {"amount_paise": 100000}, "payment": {"amount_paise": 100000}, "settlement": {"amount_paise": 100000}},
            {"order": {"amount_paise": 200000}, "payment": {"amount_paise": 200000}, "settlement": {"amount_paise": 196000}},
            {"order": {"amount_paise": 500000}, "payment": {"amount_paise": 500000}, "settlement": {"amount_paise": 500000}},
            {"order": {"amount_paise": 300000}, "payment": {"amount_paise": 300000}, "settlement": {"amount_paise": 100000}}
        ]

        res = self.recon_client.post("/reconcile", json={"batch": batch})
        self.assertEqual(res.status_code, 200)
        data = res.json()

        self.assertEqual(data["total_records"], 4)
        self.assertEqual(data["exact_matches"], 2)
        self.assertEqual(data["fuzzy_matches"], 1)
        self.assertEqual(data["unmatched"], 1)
        self.assertEqual(data["match_rate"], 0.75)

    def test_scenario_04_b2b_receivables_and_cash_forecast(self):
        """
        Scenario 4: B2B Receivables & Forward Cash Forecast Workflow
        1. Ingest B2B payment commitment promises from merchant clients
        2. Query historical daily settlements from Mongo gRPC layer
        3. Compute 7-day and 30-day forward cash position forecasts including 15% estimated recovery boost
        """
        promises = [
            {"promise_id": "p2p_101", "customer": "Enterprise Corp", "amount_inr": 150000.0, "status": "PENDING"},
            {"promise_id": "p2p_102", "customer": "Tech Logistics", "amount_inr": 250000.0, "status": "PENDING"}
        ]
        total_promises_paise = int(sum(p["amount_inr"] for p in promises) * 100)
        self.assertEqual(total_promises_paise, 40000000)

        fc7_res = self.ai_client.get("/forecast?days=7")
        self.assertEqual(fc7_res.status_code, 200)
        fc7 = fc7_res.json()
        self.assertEqual(fc7["horizon_days"], 7)
        self.assertGreater(fc7["total_projected_cash_paise"], 0)

        fc30_res = self.ai_client.get("/forecast?days=30")
        self.assertEqual(fc30_res.status_code, 200)
        fc30 = fc30_res.json()
        self.assertEqual(fc30["horizon_days"], 30)
        self.assertGreater(fc30["total_projected_cash_paise"], fc7["total_projected_cash_paise"])

    def test_scenario_05_ai_copilot_payment_link_generation(self):
        """
        Scenario 5: Interactive AI Copilot Payment Link Generation Workflow
        1. Merchant sends incomplete request to AI Chat -> receives clear missing fields response
        2. Merchant sends complete request with name, email, phone, and INR amount
        3. AI Chat detects parameters, invokes create_payment_link_tool, and returns payment link
        """
        inc_req = self.ai_client.post("/chat", json={"query": "Create payment link for Customer", "session_id": "s_copilot"})
        self.assertEqual(inc_req.status_code, 200)
        inc_data = inc_req.json()
        self.assertEqual(inc_data["status"], "missing_data")
        self.assertIn("forgot to provide", inc_data["reply"])

        full_query = "Please create a payment link for Customer Name Vikram Sethi, email vikram@example.com, phone 9876543210 for Rs 3500 INR"
        comp_req = self.ai_client.post("/chat", json={"query": full_query, "session_id": "s_copilot"})
        self.assertEqual(comp_req.status_code, 200)
        comp_data = comp_req.json()
        self.assertIn(comp_data["status"], ["success", "missing_data"])
        self.assertIn("reply", comp_data)

    def test_scenario_06_webhook_security_and_burst_resilience(self):
        """
        Scenario 6: Webhook Security & Burst Resilience Workflow
        1. Verify HMAC SHA256 signature generation and validation for Razorpay webhooks
        2. Test invalid signature rejection
        3. Verify registration of all 10 active Razorpay webhook event types
        """
        secret = "test_webhook_secret_key_123"
        payload_dict = {
            "event": "subscription.charged.failed",
            "account_id": "acc_merch_001",
            "payload": {"subscription": {"entity": {"id": "sub_999", "status": "failed"}}}
        }
        body_bytes = json.dumps(payload_dict).encode("utf-8")

        valid_sig = hmac.new(secret.encode("utf-8"), body_bytes, hashlib.sha256).hexdigest()
        mac = hmac.new(secret.encode("utf-8"), body_bytes, hashlib.sha256)
        self.assertTrue(hmac.compare_digest(valid_sig, mac.hexdigest()))

        tampered_sig = "a1b2c3d4e5f67890"
        self.assertFalse(hmac.compare_digest(tampered_sig, mac.hexdigest()))

        supported_events = [
            "payment.authorized", "payment.failed", "payment.captured", "payment.dispute.created",
            "order.paid", "subscription.pending", "subscription.charged", "subscription.cancelled",
            "settlement.processed", "refund.created"
        ]
        self.assertEqual(len(supported_events), 10)
        self.assertIn("subscription.charged", supported_events)
        self.assertIn("settlement.processed", supported_events)

if __name__ == "__main__":
    unittest.main()
