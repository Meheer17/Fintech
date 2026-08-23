import os
import sys
import pytest

# Add all microservice source directories to sys.path
BASE_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

sys.path.insert(0, os.path.join(BASE_DIR, "ai_gateway", "src"))
sys.path.insert(0, os.path.join(BASE_DIR, "failure_detector", "src"))
sys.path.insert(0, os.path.join(BASE_DIR, "recovery_orchestrator", "src"))
sys.path.insert(0, os.path.join(BASE_DIR, "reconciliation_engine", "src"))
sys.path.insert(0, os.path.join(BASE_DIR, "data"))

@pytest.fixture(scope="session")
def base_dir():
    return BASE_DIR
