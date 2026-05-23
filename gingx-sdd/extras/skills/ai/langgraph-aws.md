---
name: langgraph-aws
description: LangGraph deploy on AWS — Bedrock (Claude), Lambda, ECS Fargate, DynamoDB checkpointer, API Gateway WebSocket streaming
category: ai
stack: [langgraph, aws, bedrock, lambda, ecs, dynamodb, api-gateway, fastapi, docker]
triggers: [aws, bedrock, lambda, ecs, dynamodb, api gateway, fargate, step functions, cloudformation, cdk, iam, secrets manager]
---

# LangGraph on AWS Skill

## Agent Attitude
Eres un arquitecto de agentes IA en AWS. Bedrock para modelos (sin API keys externas).
DynamoDB para checkpoint (serverless). ECS Fargate para hosting de grafo.
Secrets Manager para toda credencial. IAM least privilege siempre.
CDK o CloudFormation. NUNCA click-ops (console manual).

## ¿Por qué AWS y no LangGraph Cloud?

| Escenario | Recomendación |
|---|---|
| Prototipo / MVP | LangGraph Cloud (más rápido) |
| Producción con compliance | AWS (VPC, IAM, audit, KMS) |
| Datos sensibles (banca, salud) | AWS (no salen de tu VPC) |
| +50K invocaciones/día | AWS (costos predecibles) |

## Arquitectura de Referencia

```
┌──────────────────────────────────────────────────────┐
│  API Gateway WebSocket                                │
│  /agent/stream  →  streaming bidireccional           │
├──────────────────────────────────────────────────────┤
│  ECS Fargate (LangGraph Server)                       │
│  ├── FastAPI + graph.compile()                       │
│  ├── checkpointer → DynamoDB                         │
│  └── secrets → Secrets Manager                       │
├──────────────────────────────────────────────────────┤
│  Amazon Bedrock                                       │
│  ├── Claude Opus 4.7  (razonamiento complejo)        │
│  ├── Claude Sonnet 4.6 (planificación, código)       │
│  └── Claude Haiku 4.5  (tools rápidas, clasificación)│
├──────────────────────────────────────────────────────┤
│  DynamoDB                                             │
│  ├── langgraph_checkpoints  (estado de sesiones)     │
│  └── langgraph_store       (memoria persistente)     │
└──────────────────────────────────────────────────────┘
```

## Bedrock (Modelos)

```python
from langchain_aws import ChatBedrock

# Claude en Bedrock — sin API key de Anthropic
model = ChatBedrock(
    model_id="anthropic.claude-opus-4-7",
    region_name="us-east-1",
    temperature=0.3,
    max_tokens=4096,
)

# Fallback entre modelos si uno falla
haiku = ChatBedrock(model_id="anthropic.claude-haiku-4-5")
sonnet = ChatBedrock(model_id="anthropic.claude-sonnet-4-6")
opus = ChatBedrock(model_id="anthropic.claude-opus-4-7")

# Router por complejidad de tarea
def choose_model(state):
    if state["complexity"] == "low": return haiku
    if state["complexity"] == "medium": return sonnet
    return opus
```

### Permisos IAM para Bedrock
```json
{
  "Effect": "Allow",
  "Action": [
    "bedrock:InvokeModel",
    "bedrock:InvokeModelWithResponseStream"
  ],
  "Resource": [
    "arn:aws:bedrock:*::foundation-model/anthropic.claude-*"
  ]
}
```

## DynamoDB (Checkpointer)

```python
from langgraph.checkpoint.dynamodb import DynamoDBSaver

checkpointer = DynamoDBSaver(
    table_name="langgraph_checkpoints",
    region_name="us-east-1",
    # TTL automático para limpiar sesiones viejas
    ttl_attribute="expires_at",
    ttl_days=7,
)

graph = workflow.compile(checkpointer=checkpointer)
```

### Tabla DynamoDB (CDK)
```python
from aws_cdk import aws_dynamodb as dynamodb

table = dynamodb.Table(self, "Checkpoints",
    table_name="langgraph_checkpoints",
    partition_key=dynamodb.Attribute(name="pk", type=dynamodb.AttributeType.STRING),
    sort_key=dynamodb.Attribute(name="sk", type=dynamodb.AttributeType.STRING),
    billing_mode=dynamodb.BillingMode.PAY_PER_REQUEST,
    time_to_live_attribute="expires_at",
    removal_policy=RemovalPolicy.RETAIN,  # NUNCA destruir en prod
)
```

## ECS Fargate (Hosting)

```dockerfile
# Dockerfile para LangGraph en ECS
FROM python:3.12-slim

WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY agent/ ./agent/

EXPOSE 8000
CMD ["uvicorn", "agent.server:app", "--host", "0.0.0.0", "--port", "8000"]
```

```python
# agent/server.py — FastAPI para ECS
from fastapi import FastAPI, Request
from fastapi.responses import StreamingResponse
from agent.graph import graph
import json

app = FastAPI()

@app.post("/agent/invoke")
async def invoke(request: Request):
    body = await request.json()
    result = graph.invoke(
        body["state"],
        config={"configurable": {"thread_id": body.get("thread_id", "default")}}
    )
    return {"result": result}

@app.post("/agent/stream")
async def stream(request: Request):
    body = await request.json()
    async def event_generator():
        for event in graph.stream(
            body["state"],
            config={"configurable": {"thread_id": body.get("thread_id")}},
            stream_mode="updates"
        ):
            yield f"data: {json.dumps(event, default=str)}\n\n"
    
    return StreamingResponse(event_generator(), media_type="text/event-stream")
```

## Secrets Manager

```python
import boto3
import json

def get_secret(secret_name: str) -> dict:
    """Recupera secretos de AWS Secrets Manager."""
    client = boto3.client("secretsmanager", region_name="us-east-1")
    response = client.get_secret_value(SecretId=secret_name)
    return json.loads(response["SecretString"])

# Uso
db_creds = get_secret("langgraph/dynamodb")
# NUNCA: DB_PASSWORD = "hardcodeado"
```

## CDK (Infraestructura como Código)

```python
from aws_cdk import Stack
from aws_cdk import aws_ecs_patterns, aws_ecs, aws_ec2

class LangGraphStack(Stack):
    def __init__(self, scope, id, **kwargs):
        super().__init__(scope, id, **kwargs)

        # Fargate Service
        cluster = aws_ecs.Cluster(self, "LangGraphCluster")

        fargate = aws_ecs_patterns.ApplicationLoadBalancedFargateService(
            self, "LangGraphService",
            cluster=cluster,
            cpu=1024,        # 1 vCPU
            memory_limit_mib=2048,  # 2 GB RAM
            desired_count=2,  # Mínimo 2 para HA
            task_image_options={
                "image": aws_ecs.ContainerImage.from_asset("."),
                "container_port": 8000,
                "environment": {
                    "BEDROCK_REGION": "us-east-1",
                    "CHECKPOINT_TABLE": "langgraph_checkpoints",
                    "LOG_LEVEL": "INFO",
                }
            },
            public_load_balancer=False,  # Internal ALB
        )

        # Auto-scaling basado en CPU
        scaling = fargate.service.auto_scale_task_count(
            min_capacity=2,
            max_capacity=20,
        )
        scaling.scale_on_cpu_utilization(
            "CpuScaling",
            target_utilization_percent=70,
        )
```

## Do's

### Networking
- VPC privada para LangGraph. NUNCA subred pública.
- VPC Endpoints para Bedrock, DynamoDB, Secrets Manager (sin internet).
- ALB interno. API Gateway como entry point público.

### Seguridad
- Secrets Manager para TODA credencial.
- KMS para encriptar checkpoints en DynamoDB.
- IAM roles por tarea (task role), no por servicio.
- CloudTrail habilitado para auditoría de todas las llamadas a Bedrock.
- WAF en API Gateway para proteger contra inyección y DDoS.

### Observabilidad
- CloudWatch Logs para todos los contenedores.
- X-Ray para tracing distribuido (API Gateway → ECS → Bedrock → DynamoDB).
- Alarma CloudWatch si errores > 5% en 5 minutos.
- LangFuse como capa adicional de tracing de agentes.

### Costos
- DynamoDB on-demand (no provisioned) para desarrollo.
- ECS Fargate Spot para ambientes no productivos.
- Bedrock on-demand (sin throughput provisioned hasta +100K invocaciones/día).

## Don'ts
- NO API key de Anthropic en Lambda/ECS. Usar Bedrock.
- NO DynamoDB provisioned en desarrollo (costo fijo sin uso).
- NO secrets en variables de entorno del task definition. Usar Secrets Manager.
- NO IPs públicas en ECS. Todo detrás de ALB interno.
- NO `AdministratorAccess` en IAM. Least privilege por nodo del grafo.
- NO `aws configure` con credenciales de largo plazo. Usar IAM roles.
- NO subir el Dockerfile con `.env` al registro de contenedores.

## Step Functions (Orquestación Híbrida)

Para pipelines que mezclan agentes IA + pasos tradicionales:

```python
# Step Function que orquesta:
# 1. LangGraph genera código (Bedrock)
# 2. CodeBuild compila
# 3. LangGraph revisa el build output
# 4. Aprobación humana (si hay error)
# 5. Deploy a ECS
```

## Recomendaciones de Modelo por Tarea

| Tarea | Modelo Bedrock | Costo aprox |
|---|---|---|
| Spec / Planificación | Claude Opus 4.7 | $$$ |
| Generación de código | Claude Sonnet 4.6 | $$ |
| Tools / Clasificación | Claude Haiku 4.5 | $ |
| Embeddings | Titan Embeddings v2 | ¢ |

## Recommended Commands
- `cdk deploy LangGraphStack` — Deploy infraestructura
- `cdk diff` — Ver cambios antes de deploy
- `aws logs tail /ecs/langgraph --follow` — Logs en tiempo real
- `aws bedrock get-foundation-model --model-id anthropic.claude-opus-4-7` — Verificar disponibilidad
- `aws dynamodb describe-table --table-name langgraph_checkpoints` — Estado de tabla
- `aws ecs update-service --cluster LangGraph --service LangGraph --force-new-deployment` — Redeploy
