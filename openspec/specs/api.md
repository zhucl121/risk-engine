# Spec: API Layer

**Status**: Active  
**Last updated**: 2026-02

## Endpoints

### POST /api/v1/decision

**Purpose**: Synchronous risk evaluation.

**Request body**:
```json
{
  "request_id": "optional-uuid-v4",
  "scene_code": "PAYMENT_CHECKOUT",
  "user_id": "u123456",
  "device_id": "d-abc-def",
  "session_id": "s-xyz",
  "ip": "1.2.3.4",
  "amount": 9900,
  "extra": {"merchant_id": "m001"}
}
```

**Constraints**:
- `scene_code` required; returns 400 if missing
- `request_id` generated (UUID v4) if not provided
- `amount` in smallest currency unit (cents); 0 for non-monetary scenes

**Response 200**:
```json
{
  "request_id": "01HZ...",
  "decision": "PASS",
  "risk_score": 120,
  "risk_level": "LOW",
  "hit_rules": [],
  "model_scores": {"payment_fraud_xgb": 0.08},
  "risk_reasons": [],
  "actions": [],
  "cost_ms": 23
}
```

**Error responses**:
| Code | Condition |
|------|-----------|
| 400 | Missing required field or malformed JSON |
| 500 | Internal engine failure (never leak stack trace) |
| 503 | Engine unhealthy (returned by health check, not decision) |

### GET /api/v1/health

**Purpose**: Liveness + readiness check.

**Response 200** (healthy):
```json
{"healthy": true, "components": {"redis": true, "kafka": true}}
```

**Response 503** (unhealthy):
```json
{"healthy": false, "components": {"redis": false, "kafka": true}}
```

### GET /metrics

Prometheus metrics endpoint (standard format). Exposed by `client_golang`.

## gRPC API

Defined in `api/grpc/proto/decision.proto`.

Service: `DecisionService`
- `Evaluate(DecisionRequest) → DecisionResponse`
- `BatchEvaluate(BatchDecisionRequest) → BatchDecisionResponse`
- `Health(HealthRequest) → HealthResponse`

## Non-Functional Requirements

| Metric | Target |
|--------|--------|
| HTTP handler overhead | < 2ms (excluding engine time) |
| Max request body size | 64KB |
| Timeout (server-side) | 100ms write timeout |
| TLS | Required in production; optional in dev |

## Versioning

- URL-versioned: `/api/v1/`
- Breaking changes require `/api/v2/`
- Proto changes: backward-compatible only within `v1`; breaking = new service version
