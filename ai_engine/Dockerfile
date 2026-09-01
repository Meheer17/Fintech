FROM python:3.12-slim

WORKDIR /app

COPY ai_engine/pyproject.toml ./
RUN pip install --no-cache-dir --upgrade pip setuptools && \
    pip install --no-cache-dir fastapi uvicorn pydantic python-dotenv requests httpx openai strands-agents strands-agents-tools "grpcio>=1.60.0,<1.70.0" "grpcio-tools>=1.60.0,<1.70.0"

COPY revenueiq_dev_kit ./revenueiq_dev_kit
COPY ai_engine ./ai_engine

WORKDIR /app/ai_engine
RUN python -m grpc_tools.protoc -I/app/revenueiq_dev_kit/proto --python_out=src/proto --grpc_python_out=src/proto /app/revenueiq_dev_kit/proto/mongo_service/mongodb_service.proto || true

EXPOSE 8006 50002 50003 50004 50006
CMD ["python", "src/main.py"]
