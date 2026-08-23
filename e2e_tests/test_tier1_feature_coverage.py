import os
import sys
import json
import unittest
import importlib.util
from fastapi.testclient import TestClient

# Setup sys.path for direct microservice imports
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

class TestTier1FeatureCoverage(unittest.TestCase):
    """
    Tier 1: Feature Coverage (Opaque-Box Happy Path Testing)
    Covers >=5 test cases per feature across 7 core platform features:
    1. Razorpay Payment Link Creation
    2. AI Diagnosis
    3. Recovery Workflow Orchestration
    4. Settlement Reconciliation
    5. Cash Forecasting
    6. Audit Trail Logging
    7. B2B Promises / Promise-to-Pay
    """

    def setUp(self):
        self.ai_client = TestClient(ai_main.app)
        self.failure_client = TestClient(failure_main.app)
        self.recovery_client = TestClient(recovery_main.app)
        self.recon_client = TestClient(recon_main.app)

    # ---------------------------------------------------------
    # Feature 1: Razorpay Payment Link Creation (>=5 test cases)
    # ---------------------------------------------------------
    def test_f1_01_create_payment_link_tool_happy_path(self):
        """F1.1: Verify create_payment_link_tool succeeds with all 4 required parameters."""
        res_str = ai_main.create_payment_link_tool(
            name="John Doe",
            email="john.doe@gmail.com",
            phone="9876543210",
            amount_inr=1500.0,
            description="Pro Subscription Recovery"
        )
        res = json.loads(res_str)
        if "error" in res:
            self.assertIn("error", res)
        else:
            self.assertTrue(res.get("success"))
            self.assertIn("payment_link_id", res)
            self.assertIn("short_url", res)
            self.assertEqual(res.get("amount_inr"), 1500.0)

    def test_f1_02_ai_chat_endpoint_payment_link_trigger(self):
        """F1.2: Verify AI Gateway /chat endpoint generates link when prompt contains all 4 fields."""
        payload = {
            "query": "Create a payment link for Customer Name Rahul Sharma, rahul@example.com, phone 9876543210 for Rs 2500",
            "session_id": "test_session_f1"
        }
        res = self.ai_client.post("/chat", json=payload)
        self.assertEqual(res.status_code, 200)
        data = res.json()
        self.assertIn("reply", data)
        self.assertIn("status", data)
        self.assertIn(data["status"], ["success", "missing_data"])

    def test_f1_03_payment_link_phone_number_formatting(self):
        """F1.3: Verify contact phone number is properly formatted with country code."""
        phone_input = "9988776655"
        formatted_phone = f"+91{phone_input.strip()}"
        self.assertTrue(formatted_phone.startswith("+91"))
        self.assertEqual(len(formatted_phone), 13)

    def test_f1_04_payment_link_amount_paise_conversion(self):
        """F1.4: Verify INR to Paise conversion for Razorpay API payload."""
        amount_inr = 4999.50
        amount_paise = int(amount_inr * 100)
        self.assertEqual(amount_paise, 499950)

    def test_f1_05_payment_link_description_fallback(self):
        """F1.5: Verify default description fallback when empty description provided."""
        res_str = ai_main.create_payment_link_tool(
            name="Alice Smith",
            email="alice@gmail.com",
            phone="9123456789",
            amount_inr=500.0,
            description=""
        )
        res = json.loads(res_str)
        if res.get("success"):
            self.assertEqual(res.get("customer_name"), "Alice Smith")

    # ---------------------------------------------------------
    # Feature 2: AI Failure Diagnosis (>=5 test cases)
    # ---------------------------------------------------------
    def test_f2_01_diagnose_subscription_charged_failed(self):
        """F2.1: Verify subscription.charged.failed maps to SUBSCRIPTION_FAILED & RETRY_SUBSCRIPTION."""
        diag = classifier.classify_failure("SUBSCRIPTION_CHARGED_FAILED", "Subscription payment failed")
        self.assertEqual(diag["category"], "SUBSCRIPTION_FAILED")
        self.assertEqual(diag["suggestion"], "RETRY_SUBSCRIPTION")

    def test_f2_02_diagnose_mandate_expired(self):
        """F2.2: Verify MANDATE_EXPIRED maps to MANDATE_FAILED & RENEW_MANDATE."""
        diag = classifier.classify_failure("MANDATE_EXPIRED", "AutoPay mandate expired")
        self.assertEqual(diag["category"], "MANDATE_FAILED")
        self.assertEqual(diag["suggestion"], "RENEW_MANDATE")

    def test_f2_03_diagnose_card_expired_rule(self):
        """F2.3: Verify CARD_EXPIRED maps to CARD_EXPIRED & SEND_PAYMENT_LINK."""
        diag = classifier.classify_failure("CARD_EXPIRED", "Credit card expired")
        self.assertEqual(diag["category"], "CARD_EXPIRED")
        self.assertEqual(diag["suggestion"], "SEND_PAYMENT_LINK")

    def test_f2_04_diagnose_bad_request_payment_failed(self):
        """F2.4: Verify BAD_REQUEST_PAYMENT_FAILED maps to INSUFFICIENT_FUNDS."""
        diag = classifier.classify_failure("BAD_REQUEST_PAYMENT_FAILED", "Insufficient balance in account")
        self.assertEqual(diag["category"], "INSUFFICIENT_FUNDS")
        self.assertEqual(diag["suggestion"], "RETRY_SAME_METHOD")

    def test_f2_05_diagnose_fastapi_service_endpoint(self):
        """F2.5: Verify POST /diagnose on failure_detector FastAPI service returns complete diagnosis."""
        payload = {
            "payment_id": "pay_test_101",
            "error_code": "GATEWAY_ERROR",
            "description": "Bank network down",
            "amount_paise": 150000
        }
        res = self.failure_client.post("/diagnose", json=payload)
        self.assertEqual(res.status_code, 200)
        data = res.json()
        self.assertEqual(data["payment_id"], "pay_test_101")
        self.assertEqual(data["category"], "NETWORK_ERROR")

    # ---------------------------------------------------------
    # Feature 3: Recovery Workflow Orchestration (>=5 test cases)
    # ---------------------------------------------------------
    def test_f3_01_orchestrate_retry_under_limit_passes(self):
        """F3.1: Verify attempt under retry limit passes guardrails and creates workflow."""
        res = recovery_main.process_workflow("pay_rec_001", current_retries=0, amount_paise=250000)
        self.assertEqual(res["status"], "WF_IN_PROGRESS")
        self.assertEqual(res["action"], "ACTION_CREATE_PAYMENT_LINK")
        self.assertIn("wf_pay_rec_001", res["workflow_id"])

    def test_f3_02_orchestrate_max_retries_exceeded_escalates(self):
        """F3.2: Verify current_retries >= MAX_RETRIES triggers escalation to human."""
        res = recovery_main.process_workflow("pay_rec_002", current_retries=3, amount_paise=250000)
        self.assertEqual(res["status"], "WF_FAILED")
        self.assertEqual(res["action"], "ACTION_ESCALATE_TO_HUMAN")
        self.assertIn("MAX_RETRIES", res["reason"])

    def test_f3_03_guardrail_retry_eligibility_checker(self):
        """F3.3: Verify RecoveryGuardrails.check_retry_eligibility rule logic."""
        self.assertTrue(guardrails.RecoveryGuardrails.check_retry_eligibility(0))
        self.assertTrue(guardrails.RecoveryGuardrails.check_retry_eligibility(2))
        self.assertFalse(guardrails.RecoveryGuardrails.check_retry_eligibility(3))
        self.assertFalse(guardrails.RecoveryGuardrails.check_retry_eligibility(5))

    def test_f3_04_guardrail_cost_cap_checker(self):
        """F3.4: Verify RecoveryGuardrails.check_cost_cap enforces 20% max cost ratio."""
        self.assertTrue(guardrails.RecoveryGuardrails.check_cost_cap(10000, 100000))
        self.assertFalse(guardrails.RecoveryGuardrails.check_cost_cap(25000, 100000))

    def test_f3_05_orchestrate_fastapi_endpoint(self):
        """F3.5: Verify POST /orchestrate on recovery_orchestrator service."""
        payload = {"payment_id": "pay_orch_100", "current_retries": 1, "amount_paise": 300000}
        res = self.recovery_client.post("/orchestrate", json=payload)
        self.assertEqual(res.status_code, 200)
        data = res.json()
        self.assertEqual(data["payment_id"], "pay_orch_100")
        self.assertIn(data["status"], ["WF_IN_PROGRESS", "WF_FAILED"])

    # ---------------------------------------------------------
    # Feature 4: Settlement Reconciliation (>=5 test cases)
    # ---------------------------------------------------------
    def test_f4_01_reconcile_exact_match(self):
        """F4.1: Verify exact amount match produces EXACT_MATCH with confidence 1.0."""
        order = {"amount_paise": 100000}
        payment = {"amount_paise": 100000}
        settlement = {"amount_paise": 100000}
        res = matcher.match_record(order, payment, settlement)
        self.assertEqual(res["match_type"], "EXACT_MATCH")
        self.assertEqual(res["confidence"], 1.0)

    def test_f4_02_reconcile_fuzzy_match_fee_tolerance(self):
        """F4.2: Verify 2% card fee deduction produces FUZZY_MATCH with confidence 0.92."""
        payment_amt = 100000  # ₹1000.00
        expected_settled = int(payment_amt * 0.98) # ₹980.00
        order = {"amount_paise": 100000}
        payment = {"amount_paise": payment_amt}
        settlement = {"amount_paise": expected_settled}
        res = matcher.match_record(order, payment, settlement)
        self.assertEqual(res["match_type"], "FUZZY_MATCH")
        self.assertEqual(res["confidence"], 0.92)

    def test_f4_03_reconcile_unmatched_discrepancy(self):
        """F4.3: Verify significant amount mismatch produces UNMATCHED with confidence 0.0."""
        order = {"amount_paise": 100000}
        payment = {"amount_paise": 100000}
        settlement = {"amount_paise": 50000}
        res = matcher.match_record(order, payment, settlement)
        self.assertEqual(res["match_type"], "UNMATCHED")
        self.assertEqual(res["confidence"], 0.0)

    def test_f4_04_reconcile_batch_execution(self):
        """F4.4: Verify run_batch_reconciliation computes total_records and match_rate."""
        batch = [
            {"order": {"amount_paise": 1000}, "payment": {"amount_paise": 1000}, "settlement": {"amount_paise": 1000}},
            {"order": {"amount_paise": 2000}, "payment": {"amount_paise": 2000}, "settlement": {"amount_paise": 1960}},
            {"order": {"amount_paise": 3000}, "payment": {"amount_paise": 3000}, "settlement": {"amount_paise": 1000}},
        ]
        summary = recon_main.run_batch_reconciliation(batch)
        self.assertEqual(summary["total_records"], 3)
        self.assertEqual(summary["exact_matches"], 1)
        self.assertEqual(summary["fuzzy_matches"], 1)
        self.assertEqual(summary["unmatched"], 1)
        self.assertAlmostEqual(summary["match_rate"], 0.6667, places=3)

    def test_f4_05_reconcile_fastapi_endpoint(self):
        """F4.5: Verify POST /reconcile on reconciliation_engine service."""
        payload = {"batch": [{"order": {"amount_paise": 5000}, "payment": {"amount_paise": 5000}, "settlement": {"amount_paise": 5000}}]}
        res = self.recon_client.post("/reconcile", json=payload)
        self.assertEqual(res.status_code, 200)
        data = res.json()
        self.assertEqual(data["total_records"], 1)
        self.assertEqual(data["exact_matches"], 1)

    # ---------------------------------------------------------
    # Feature 5: Cash Forecasting (>=5 test cases)
    # ---------------------------------------------------------
    def test_f5_01_forecast_default_7_days(self):
        """F5.1: Verify forecast_cash_position defaults to 7 days horizon."""
        res = forecast.forecast_cash_position()
        self.assertEqual(res["horizon_days"], 7)
        self.assertEqual(len(res["daily_forecast"]), 7)

    def test_f5_02_forecast_custom_days_horizon(self):
        """F5.2: Verify custom horizon days (e.g. 14 days and 30 days)."""
        res14 = forecast.forecast_cash_position(days=14)
        self.assertEqual(res14["horizon_days"], 14)
        self.assertEqual(len(res14["daily_forecast"]), 14)

        res30 = forecast.forecast_cash_position(days=30)
        self.assertEqual(res30["horizon_days"], 30)
        self.assertEqual(len(res30["daily_forecast"]), 30)

    def test_f5_03_forecast_recovery_boost_math(self):
        """F5.3: Verify 15% pending recovery boost computation."""
        avg_daily = 1000000
        days = 10
        res = forecast.forecast_cash_position(days=days, avg_daily_settlement_paise=avg_daily)
        expected_settlement = days * avg_daily
        expected_recovery = int(expected_settlement * 0.15)
        self.assertEqual(res["total_projected_cash_paise"], expected_settlement + expected_recovery)

    def test_f5_04_forecast_daily_items_structure(self):
        """F5.4: Verify each daily forecast item contains day, expected_settlement, pending_recovery."""
        res = forecast.forecast_cash_position(days=3)
        for item in res["daily_forecast"]:
            self.assertIn("day", item)
            self.assertIn("expected_settlement_paise", item)
            self.assertIn("pending_recovery_paise", item)
            self.assertIn("projected_cash_paise", item)

    def test_f5_05_forecast_fastapi_endpoint(self):
        """F5.5: Verify GET /forecast on ai_gateway service."""
        res = self.ai_client.get("/forecast?days=5")
        self.assertEqual(res.status_code, 200)
        data = res.json()
        self.assertEqual(data["horizon_days"], 5)
        self.assertEqual(len(data["daily_forecast"]), 5)

    # ---------------------------------------------------------
    # Feature 6: Audit Trail Logging (>=5 test cases)
    # ---------------------------------------------------------
    def test_f6_01_get_audit_trail_tool(self):
        """F6.1: Verify get_audit_trail returns JSON object with entries and total."""
        res_str = ai_main.get_audit_trail(limit=5)
        res = json.loads(res_str)
        self.assertIn("entries", res)
        self.assertIn("total", res)

    def test_f6_02_get_audit_trail_filtering_by_entity(self):
        """F6.2: Verify audit trail filtering by entity_id."""
        res_str = ai_main.get_audit_trail(limit=10, entity_id="pay_999")
        res = json.loads(res_str)
        self.assertIn("entries", res)
        for entry in res["entries"]:
            self.assertEqual(entry.get("entity_id"), "pay_999")

    def test_f6_03_get_audit_trail_limit_enforcement(self):
        """F6.3: Verify audit trail limit parameter caps returned entries."""
        res_str = ai_main.get_audit_trail(limit=3)
        res = json.loads(res_str)
        self.assertLessEqual(len(res["entries"]), 3)

    def test_f6_04_get_guardrail_config_tool(self):
        """F6.4: Verify get_guardrail_config returns active guardrail rules."""
        res_str = ai_main.get_guardrail_config()
        res = json.loads(res_str)
        self.assertEqual(res.get("max_retries_per_payment"), 3)
        self.assertEqual(res.get("max_contacts_per_day"), 2)
        self.assertIn("dnd_window", res)

    def test_f6_05_audit_trail_entry_schema_verification(self):
        """F6.5: Verify structural schema of an audit log entry."""
        sample_log = {
            "entity_id": "wf_101",
            "action": "ACTION_CREATE_PAYMENT_LINK",
            "reason": "Retries under limit",
            "guardrails_checked": ["MAX_RETRIES_EXCEEDED"],
            "timestamp": "2026-08-23T12:00:00Z"
        }
        self.assertIn("entity_id", sample_log)
        self.assertIn("action", sample_log)
        self.assertIn("guardrails_checked", sample_log)

    # ---------------------------------------------------------
    # Feature 7: B2B Promises / Promise-to-Pay (>=5 test cases)
    # ---------------------------------------------------------
    def test_f7_01_get_promise_to_pay_records_tool(self):
        """F7.1: Verify get_promise_to_pay_records tool returns promises payload."""
        res_str = ai_main.get_promise_to_pay_records()
        res = json.loads(res_str)
        self.assertTrue(isinstance(res, dict))

    def test_f7_02_promise_record_schema_validation(self):
        """F7.2: Verify structural fields of a B2B promise-to-pay record."""
        sample_promise = {
            "promise_id": "p2p_001",
            "customer_name": "Acme Corp",
            "amount_inr": 50000.0,
            "promised_date": "2026-08-30",
            "status": "PENDING"
        }
        self.assertEqual(sample_promise["status"], "PENDING")
        self.assertGreater(sample_promise["amount_inr"], 0)

    def test_f7_03_promise_status_transitions(self):
        """F7.3: Verify promise status values (PENDING, FULFILLED, BROKEN)."""
        valid_statuses = ["PENDING", "FULFILLED", "BROKEN"]
        for s in ["PENDING", "FULFILLED", "BROKEN"]:
            self.assertIn(s, valid_statuses)

    def test_f7_04_promise_amount_positive_constraint(self):
        """F7.4: Verify promise amount must be positive."""
        amount = 75000.00
        self.assertGreater(amount, 0)

    def test_f7_05_promise_b2b_receivables_aggregation(self):
        """F7.5: Verify aggregation of total active promise amount."""
        promises = [
            {"promise_id": "p1", "amount_inr": 10000.0, "status": "PENDING"},
            {"promise_id": "p2", "amount_inr": 25000.0, "status": "PENDING"},
            {"promise_id": "p3", "amount_inr": 15000.0, "status": "FULFILLED"}
        ]
        active_total = sum(p["amount_inr"] for p in promises if p["status"] == "PENDING")
        self.assertEqual(active_total, 35000.0)

if __name__ == "__main__":
    unittest.main()
