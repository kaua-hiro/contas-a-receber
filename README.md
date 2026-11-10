# Contas a Receber · Go + Vue

![CI](https://github.com/kaua-hiro/contas-a-receber/actions/workflows/ci.yml/badge.svg)
![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)
![Vue](https://img.shields.io/badge/Vue-3-42b883?logo=vuedotjs&logoColor=white)
![License](https://img.shields.io/badge/license-MIT-blue)

CRUD de **títulos a receber** (cobranças): uma API REST em Go, só com a biblioteca padrão, e uma tela em Vue 3 para cadastrar, filtrar, editar, dar baixa e acompanhar os totais.

## Funcionalidades

- CRUD completo de títulos (cliente, valor, vencimento, status)
- **Dar baixa** em um título (`PATCH /titulos/{id}/pagar`), com `409 Conflict` se ele já estiver pago
- Filtros por **status** e **cliente** (busca parcial, sem diferenciar maiúsculas)
- **Resumo financeiro**: total em aberto, total recebido e títulos vencidos
- Validação de entrada com mensagens claras e status HTTP corretos (`400`, `404`, `409`, `422`)
- Destaque visual para títulos vencidos

## Stack

| Camada | Tecnologias |
|---|---|
| API | Go 1.22+ (`net/http`, `encoding/json`, `sync`), sem dependências externas |
| Front | Vue 3 (Composition API, `<script setup>`), Vite |
| Testes | `go test` + `httptest` (testes em tabela) |
| CI | GitHub Actions (`go vet`, `go test -race`, build do front) |

## Arquitetura

```
.
├── api/
│   ├── main.go                  # servidor, CORS, log e configuração por env
│   └── internal/titulo/
│       ├── titulo.go            # domínio: entidade, validação, erros
│       ├── store.go             # interface Repositorio + implementação em memória
│       ├── handler.go           # camada HTTP (rotas, JSON, status codes)
│       └── handler_test.go      # testes de integração dos endpoints
├── web/
│   └── src/
│       ├── api.js               # cliente HTTP e formatação (BRL, datas)
│       ├── App.vue              # listagem, filtros e resumo
│       └── components/
│           └── TituloForm.vue   # formulário de criar/editar (props + emit)
└── docs/consultas.sql           # modelo relacional e consultas SQL equivalentes
```

O fluxo é `handler → Repositorio (interface) → MemoriaStore`. O handler depende só da interface, então trocar a memória por PostgreSQL significa escrever outra implementação de `Repositorio`, sem mexer na camada HTTP.

### Decisões técnicas

- **Dinheiro em centavos (`int64`)**: `float64` gera erro de arredondamento (`0.1 + 0.2 != 0.3`). O front converte reais para centavos com `Math.round(valor * 100)`.
- **`sync.RWMutex` no store**: cada requisição roda em uma goroutine, e o lock evita *race condition* no `map`. Leituras usam `RLock` e podem acontecer em paralelo.
- **Erros de domínio com `errors.Is`**: o store retorna `ErrNaoEncontrado` / `ErrJaPago`, e o handler traduz para `404` / `409` em um único lugar.
- **Entrada segura**: `DisallowUnknownFields`, limite de 1 MB no corpo (`MaxBytesReader`) e `ReadHeaderTimeout` no servidor.
- **Roteamento do Go 1.22** (`"GET /titulos/{id}"` + `r.PathValue`), sem precisar de framework.
- **Relógio injetável** (`agora func() time.Time`), o que deixa determinístico o teste de "vencido".

## Como rodar

Pré-requisitos: **Go 1.22+** e **Node 20+**.

```bash
# 1. API (http://localhost:8080)
cd api
go run .

# 2. Front (http://localhost:5173), em outro terminal
cd web
npm install
npm run dev
```

Variáveis de ambiente (opcionais):

| Variável | Onde | Padrão |
|---|---|---|
| `PORT` | api | `8080` |
| `CORS_ORIGIN` | api | `http://localhost:5173` |
| `VITE_API_URL` | web (`.env`) | `http://localhost:8080` |

## Testes

```bash
cd api
go vet ./...
go test ./... -v
```

Os testes cobrem criação válida e inválida, busca, atualização, baixa duplicada, remoção, filtros e o resumo.

## API

Base: `http://localhost:8080`

| Método | Rota | Descrição | Sucesso |
|---|---|---|---|
| `GET` | `/titulos?status=&cliente=` | Lista títulos com filtros opcionais | `200` |
| `GET` | `/titulos/resumo` | Totais em aberto, recebido e vencido | `200` |
| `GET` | `/titulos/{id}` | Busca um título | `200` |
| `POST` | `/titulos` | Cria um título | `201` |
| `PUT` | `/titulos/{id}` | Atualiza um título | `200` |
| `PATCH` | `/titulos/{id}/pagar` | Dá baixa no título | `200` |
| `DELETE` | `/titulos/{id}` | Remove um título | `204` |
| `GET` | `/health` | Health check | `200` |

Erros retornam `{"erro": "mensagem"}`: `400` (JSON ou id inválido), `404` (não encontrado), `409` (já pago), `422` (validação).

### Exemplos com curl

```bash
# Criar
curl -X POST localhost:8080/titulos \
  -H "Content-Type: application/json" \
  -d '{"cliente":"Loja A","valor_centavos":15050,"vencimento":"2026-12-01"}'

# Listar só os abertos de um cliente
curl "localhost:8080/titulos?status=ABERTO&cliente=loja"

# Dar baixa
curl -X PATCH localhost:8080/titulos/1/pagar

# Resumo
curl localhost:8080/titulos/resumo
```

Resposta de exemplo:

```json
{
  "id": 1,
  "cliente": "Loja A",
  "valor_centavos": 15050,
  "vencimento": "2026-12-01",
  "status": "ABERTO",
  "criado_em": "2026-11-10T21:00:00Z"
}
```

## Próximos passos

- [ ] Persistência em PostgreSQL (o modelo está em [`docs/consultas.sql`](docs/consultas.sql))
- [ ] Paginação na listagem
- [ ] Autenticação
- [ ] Testes de componente no front (Vitest)

## Licença

[MIT](LICENSE) © Kauã Hiro Mizumoto
