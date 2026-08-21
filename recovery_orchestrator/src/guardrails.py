# RevenueIQ Hard-coded Recovery Guardrails

class RecoveryGuardrails:
    MAX_RETRIES = 3
    MAX_CONTACTS_PER_DAY = 2
    MAX_RECOVERY_COST_PERCENT = 0.20 # 20% max cost ratio
    MIN_CONTACT_HOUR_IST = 9  # 9 AM IST
    MAX_CONTACT_HOUR_IST = 21 # 9 PM IST

    @classmethod
    def check_retry_eligibility(cls, current_retries: int) -> bool:
        return current_retries < cls.MAX_RETRIES

    @classmethod
    def check_contact_limit(cls, contacts_today: int) -> bool:
        return contacts_today < cls.MAX_CONTACTS_PER_DAY

    @classmethod
    def check_cost_cap(cls, estimated_cost_paise: int, amount_paise: int) -> bool:
        if amount_paise <= 0:
            return False
        return (estimated_cost_paise / amount_paise) <= cls.MAX_RECOVERY_COST_PERCENT
