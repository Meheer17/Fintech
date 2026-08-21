# RevenueIQ Forward Cash Position Forecaster

def forecast_cash_position(days: int = 7, avg_daily_settlement_paise: int = 2500000) -> dict:
    projected_settlements = days * avg_daily_settlement_paise
    pending_recovery_estimate = int(projected_settlements * 0.15) # 15% estimated recovery boost
    net_position = projected_settlements + pending_recovery_estimate

    forecast_items = []
    for day in range(1, days + 1):
        forecast_items.append({
            "day": day,
            "expected_settlement_paise": avg_daily_settlement_paise,
            "pending_recovery_paise": int(avg_daily_settlement_paise * 0.15),
            "projected_cash_paise": int(avg_daily_settlement_paise * 1.15)
        })

    return {
        "horizon_days": days,
        "total_projected_cash_paise": net_position,
        "daily_forecast": forecast_items
    }
