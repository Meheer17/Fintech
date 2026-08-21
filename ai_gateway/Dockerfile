# =========================
# STAGE 1: BASE
# =========================
FROM python:3.12-slim AS base

ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    PIP_NO_CACHE_DIR=1 \
    PIP_DISABLE_PIP_VERSION_CHECK=1

WORKDIR /app

# =========================
# STAGE 2: BUILDER
# =========================
FROM base AS builder

RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    g++ \
    libffi-dev \
    curl \
    && rm -rf /var/lib/apt/lists/*

RUN python -m venv /opt/venv
ENV PATH="/opt/venv/bin:$PATH"

COPY ai_gateway/pyproject.toml .
RUN pip install --upgrade pip hatchling && \
    pip install --no-cache-dir fastapi uvicorn grpcio protobuf pydantic httpx python-dotenv

# =========================
# STAGE 3: DEV
# =========================
FROM base AS dev

COPY --from=builder /opt/venv /opt/venv
ENV PATH="/opt/venv/bin:$PATH"

COPY revenueiq_dev_kit/proto/mongo_service/proto/mongo_service ./src/proto/mongo_service
COPY ai_gateway/src ./src
COPY ai_gateway/pyproject.toml .

EXPOSE 50006 8006
CMD ["python", "src/main.py"]

# =========================
# STAGE 4: PROD (MINIMAL RUNTIME)
# =========================
FROM python:3.12-slim AS prod

ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    PATH="/opt/venv/bin:$PATH"

RUN groupadd -g 10001 appgroup && \
    useradd -u 10001 -g appgroup -s /bin/sh -d /app appuser

WORKDIR /app

COPY --from=builder /opt/venv /opt/venv
COPY --chown=appuser:appgroup ai_gateway/src ./src

RUN python -m compileall /app/src

USER appuser
EXPOSE 50006 8006
ENTRYPOINT ["python", "src/main.py"]
