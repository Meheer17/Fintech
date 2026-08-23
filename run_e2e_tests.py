#!/usr/bin/env python3
"""
RevenueIQ E2E Test Suite Runner Script
Executes all 4 tiers of opaque-box requirement-driven tests:
- Tier 1: Feature Coverage (>=5 test cases per feature across 7 core features)
- Tier 2: Boundary & Corner Cases (>=5 test cases per feature for edge cases)
- Tier 3: Cross-Feature Combinations (Pairwise & Multi-Feature Interactions)
- Tier 4: Real-World Application Scenarios (End-to-End Workflows)

Reports per-tier test counts, duration, and returns exit code 0 when all tests pass.
"""

import os
import sys
import time
import unittest

# Ensure repo base directory is in sys.path
REPO_DIR = os.path.dirname(os.path.abspath(__file__))
if REPO_DIR not in sys.path:
    sys.path.insert(0, REPO_DIR)
sys.path.insert(0, os.path.join(REPO_DIR, "ai_gateway", "src"))
sys.path.insert(0, os.path.join(REPO_DIR, "failure_detector", "src"))
sys.path.insert(0, os.path.join(REPO_DIR, "recovery_orchestrator", "src"))
sys.path.insert(0, os.path.join(REPO_DIR, "reconciliation_engine", "src"))

# Import test modules directly
from e2e_tests.test_tier1_feature_coverage import TestTier1FeatureCoverage
from e2e_tests.test_tier2_boundary_corner import TestTier2BoundaryCornerCases
from e2e_tests.test_tier3_cross_feature import TestTier3CrossFeatureCombinations
from e2e_tests.test_tier4_real_world_scenarios import TestTier4RealWorldScenarios
from e2e_tests.test_tier5_adversarial_hardening import TestTier5AdversarialHardening

def run_suite():
    loader = unittest.TestLoader()

    tier1_suite = loader.loadTestsFromTestCase(TestTier1FeatureCoverage)
    tier2_suite = loader.loadTestsFromTestCase(TestTier2BoundaryCornerCases)
    tier3_suite = loader.loadTestsFromTestCase(TestTier3CrossFeatureCombinations)
    tier4_suite = loader.loadTestsFromTestCase(TestTier4RealWorldScenarios)
    tier5_suite = loader.loadTestsFromTestCase(TestTier5AdversarialHardening)

    print("======================================================================")
    print("           REVENUEIQ DUAL-TRACK E2E TEST SUITE RUNNER                 ")
    print("======================================================================")
    start_time = time.time()

    runner = unittest.TextTestRunner(verbosity=1)

    print("\nExecuting Tier 1: Feature Coverage Tests...")
    t1_res = runner.run(tier1_suite)

    print("\nExecuting Tier 2: Boundary & Corner Case Tests...")
    t2_res = runner.run(tier2_suite)

    print("\nExecuting Tier 3: Cross-Feature Combination Tests...")
    t3_res = runner.run(tier3_suite)

    print("\nExecuting Tier 4: Real-World Application Scenario Tests...")
    t4_res = runner.run(tier4_suite)

    print("\nExecuting Tier 5: Adversarial Coverage Hardening Tests...")
    t5_res = runner.run(tier5_suite)

    elapsed_time = time.time() - start_time

    t1_passed = t1_res.testsRun - len(t1_res.failures) - len(t1_res.errors)
    t2_passed = t2_res.testsRun - len(t2_res.failures) - len(t2_res.errors)
    t3_passed = t3_res.testsRun - len(t3_res.failures) - len(t3_res.errors)
    t4_passed = t4_res.testsRun - len(t4_res.failures) - len(t4_res.errors)
    t5_passed = t5_res.testsRun - len(t5_res.failures) - len(t5_res.errors)

    total_tests = t1_res.testsRun + t2_res.testsRun + t3_res.testsRun + t4_res.testsRun + t5_res.testsRun
    total_passed = t1_passed + t2_passed + t3_passed + t4_passed + t5_passed
    total_failures = len(t1_res.failures) + len(t2_res.failures) + len(t3_res.failures) + len(t4_res.failures) + len(t5_res.failures)
    total_errors = len(t1_res.errors) + len(t2_res.errors) + len(t3_res.errors) + len(t4_res.errors) + len(t5_res.errors)

    print("\n======================================================================")
    print("                      E2E TEST SUMMARY REPORT                         ")
    print("======================================================================")
    print(f"  Tier 1 (Feature Coverage)        : {t1_passed}/{t1_res.testsRun} PASSED")
    print(f"  Tier 2 (Boundary & Corner Cases) : {t2_passed}/{t2_res.testsRun} PASSED")
    print(f"  Tier 3 (Cross-Feature Pairwise)  : {t3_passed}/{t3_res.testsRun} PASSED")
    print(f"  Tier 4 (Real-World E2E Scenarios): {t4_passed}/{t4_res.testsRun} PASSED")
    print(f"  Tier 5 (Adversarial Coverage)    : {t5_passed}/{t5_res.testsRun} PASSED")
    print("----------------------------------------------------------------------")
    print(f"  TOTAL EXECUTED : {total_tests}")
    print(f"  TOTAL PASSED   : {total_passed}")
    print(f"  TOTAL FAILED   : {total_failures}")
    print(f"  TOTAL ERRORS   : {total_errors}")
    print(f"  EXECUTION TIME : {elapsed_time:.2f} seconds")
    print("======================================================================")

    if total_failures == 0 and total_errors == 0:
        print("  OVERALL RESULT : 100% SUCCESS (EXIT CODE 0)")
        print("======================================================================")
        return 0
    else:
        print("  OVERALL RESULT : FAILURE DETECTED (EXIT CODE 1)")
        print("======================================================================")
        return 1

if __name__ == "__main__":
    sys.exit(run_suite())
