# Parts API — Motor de Priorização de Reposição de Estoque

Microserviço em Go para gerenciamento de peças com cálculo automático de
prioridade de reposição baseado em estoque, consumo e criticidade.

## Stack

- **Go 1.25+** — linguagem principal
- **chi/v5** — roteador HTTP leve e idiomático
- **SQLite** (`modernc.org/sqlite`, sem CGO) — banco local
- **testify** — assertions nos testes
- **slog** — structured logging nativo (zero dependência extra)

## Como rodar

```bash
make run                        # sobe em :8080 com SQLite (arquivo parts.db)
IN_MEMORY=true make run         # sobe com repositório em memória (sem disco)
PORT=3000 make run              # porta customizada
```

## Docker

```bash
docker build -t parts-api .
docker run -p 8080:8080 parts-api                          # SQLite
docker run -p 8081:8080 -e IN_MEMORY=true parts-api        # Memória
docker compose up api-sqlite                                # Docker Compose SQLite
docker compose up api-memory                                # Docker Compose Memória
```

## Como testar

```bash
make test                       # todos os testes com detector de race
make test-unit                  # só unitários (rápido, sem I/O)
make test-e2e                   # só E2E via httptest
make fuzz                       # 30s de fuzzing no cálculo de urgência
make bench                      # benchmarks (performance com até 10k peças)
make lint                       # go vet
make coverage                   # html de cobertura
```

## Endpoints

### Criar peça

```bash
curl -s -X POST http://localhost:8080/api/v1/parts \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Filtro de Óleo X",
    "category": "motor",
    "currentStock": 15,
    "minimumStock": 20,
    "averageDailySales": 4,
    "leadTimeDays": 5,
    "unitCost": 18.50,
    "criticalityLevel": 3
  }' | jq
```

- `id` é opcional — se não enviado, o servidor gera UUID v4 automaticamente
- `criticalityLevel` deve ser 1 a 5
- Resposta `201 Created`

### Listar peças

```bash
curl -s http://localhost:8080/api/v1/parts | jq
curl -s "http://localhost:8080/api/v1/parts?category=motor" | jq
curl -s "http://localhost:8080/api/v1/parts?needsRestock=true" | jq
curl -s "http://localhost:8080/api/v1/parts?page=1&pageSize=20" | jq
```

### Buscar por ID

```bash
curl -s http://localhost:8080/api/v1/parts/550e8400-e29b-41d4-a716-446655440000 | jq
```

### Atualizar peça (campos parciais)

```bash
curl -s -X PUT http://localhost:8080/api/v1/parts/550e8400-e29b-41d4-a716-446655440000 \
  -H "Content-Type: application/json" \
  -d '{"currentStock": 30, "unitCost": 22.90}' | jq
```

### Remover peça

```bash
curl -s -X DELETE http://localhost:8080/api/v1/parts/550e8400-e29b-41d4-a716-446655440000
```

### Prioridades de reposição

```bash
curl -s http://localhost:8080/api/v1/restock/priorities | jq
```

Retorna todas as peças ordenadas por urgência, com tie-break.

### Health check

```bash
curl -s http://localhost:8080/health | jq
```

## Regras de negócio

```
expectedConsumption = averageDailySales × leadTimeDays
projectedStock      = currentStock − expectedConsumption
needsRestock        = projectedStock < minimumStock
urgencyScore        = (minimumStock − projectedStock) × criticalityLevel
```

**Tie-break (mesmo urgencyScore):**
1. Maior `criticalityLevel`
2. Maior `averageDailySales`
3. Ordem alfabética por `name`

## Estrutura do projeto

```
cmd/api/main.go                     → entrypoint, graceful shutdown
internal/
  domain/                           → entidade Part + Value Objects puros
  usecase/                          → orquestração CRUD + prioridades
  repository/                       → interface PartRepository
    memory/                         → implementação em memória (testes)
    sqlite/                         → implementação SQLite (produção)
  http/                             → handlers, rotas, DTOs, middleware
test/
  unit/                             → table-driven + fuzz + benchmarks + concorrência
  e2e/                              → API real via httptest
```

## Decisões

As escolhas de arquitetura e os motivos estão em [`DECISIONS.md`](./DECISIONS.md).

O enunciado original do desafio está em [`README-TEST.md`](./README-TEST.md).
