# Mentor API - Sistema de Elaboração de Provas e Simulados

API Web RESTful desenvolvida em **Go (Golang)** seguindo os princípios de **Clean Architecture**, **SOLID**, **Clean Code**, e desacoplamento com implementação concreta de repositórios em **PostgreSQL**.

---

## 🏛️ 1. Arquitetura do Sistema (Clean Architecture)

A estrutura do projeto foi organizada em camadas concêntricas onde as dependências apontam estritamente de fora para dentro:

```
mentor-v2.1/
├── cmd/
│   ├── api/
│   │   └── main.go              # Ponto de entrada, injeção de dependências e graceful shutdown
│   ├── seeder/
│   │   └── main.go              # CLI interno para povoamento de questões e alternativas
│   └── migrate/
│       └── main.go              # CLI de migrações e controle de versão do banco
├── internal/
│   ├── domain/                  # Enterprise Business Rules (independente de frameworks externos)
│   │   ├── entity/              # Entidades: User, Question, Alternative, Exam, History, Analytics, etc.
│   │   ├── repository/          # Interfaces (Ports) dos repositórios
│   │   └── errors.go            # Erros de domínio (ErrNotFound, ErrForbidden, etc.)
│   ├── usecase/                 # Application Business Rules (casos de uso)
│   │   ├── auth/                # Registro, login e perfil do usuário
│   │   ├── exam/                # Criação, listagem, resolução e finalização de simulados
│   │   ├── history/             # Histórico de notas e tentativas
│   │   ├── question/            # Serviço interno de questões/alternativas (SEM exposição HTTP)
│   │   └── analytics/           # Análise de performance, métricas, rankings e evolução
│   ├── infrastructure/          # Frameworks, Drivers & Adapters
│   │   ├── config/              # Variáveis de ambiente e configuração
│   │   ├── database/            # Conexão PostgreSQL (pgx), migrator e embed de migrações
│   │   │   ├── migrations/      # Migrações versionadas embutidas (000001_..., 000002_...)
│   │   │   ├── migrator.go      # Sistema de controle de versão (schema_migrations)
│   │   │   └── postgres.go      # Pool de conexões pgx/stdlib
│   │   ├── repository/postgres/ # Implementações concretas dos repositórios em PostgreSQL
│   │   │   ├── exam/            # exam_repository.go, create.go, get_by_id.go, update.go, etc.
│   │   │   ├── user/            # user_repository.go, create.go, get_by_id.go, get_by_email.go, update.go
│   │   │   ├── question/        # question_repository.go, helpers.go, create.go, list_by_areas.go, etc.
│   │   │   ├── history/         # history_repository.go, create.go, get_by_id.go, list_by_user_id.go, etc.
│   │   │   └── analytics/       # analytics_repository.go, get_overview.go, get_by_discipline.go, etc.
│   │   └── security/            # Hashing seguro com Bcrypt e emissão/validação de tokens JWT
│   └── interface/http/          # Interface Adapters (HTTP)
│       ├── handler/             # Controllers HTTP (Auth, Exam, History, Analytics)
│       ├── middleware/          # Autenticação JWT, Logging estruturado, Recuperação e CORS
│       ├── response/            # Padronização de respostas JSON e mapeador de erros
│       └── router.go            # Configuração das rotas Chi v5
├── migrations/                  # Arquivos de migração versionados (executados em ordem determinística)
│   ├── 000001_init_schema.sql           # V1: Criação de tabelas e índices básicos
│   └── 000002_add_analytics_indexes.sql # V2: Índices para consultas analíticas e performance
├── tests/                       # Testes automatizados e de integração SEPARADOS do código-fonte
│   ├── mocks/                   # Implementações em memória dos repositórios para testes
│   ├── unit/                    # Testes unitários dos casos de uso, regras de negócio e migrador
│   └── integration/             # Testes de integração end-to-end das rotas HTTP
├── Dockerfile                   # Build multi-etapas minimalista em Alpine
├── docker-compose.yml           # Orquestração do banco PostgreSQL + API
├── Makefile                     # Atalhos para compilação, testes, migrações e execução
└── schema.prisma                # Esquema de dados original Prisma
```

---

## 🔒 2. Segurança dos Dados e Isolamento por Usuário

1. **Autenticação JWT**:
   - As rotas públicas são apenas `/api/v1/auth/register`, `/api/v1/auth/login` e `/health`.
   - Todas as demais rotas exigem cabeçalho `Authorization: Bearer <token>`.
2. **Isolamento de Simulados**:
   - Cada usuário só pode visualizar, listar, responder, finalizar e deletar os **seus próprios simulados**.
   - Qualquer tentativa de um usuário de acessar o simulado de outro usuário é bloqueada na camada de caso de uso com `domain.ErrForbidden` (HTTP 403 Forbidden).
3. **Proteção Contra Vazamento de Gabarito**:
   - Enquanto o simulado estiver em andamento, o campo `is_correct` das alternativas **não é exposto** no payload JSON.
   - O gabarito e a validação (`got_right`) só são revelados após a finalização formal do exame (`/finish`).

---

## 🚫 3. Isolamento de Questões e Alternativas (Regra de Negócio)

Conforme os requisitos da especificação:
> *"A manipulação de questões e alternativas deve ser feita somente dentro do sistema, utilizando serviços/usecases onde necessário, não expondo através de endpoints http."*

- **Não existem rotas HTTP** para CRUD de questões ou alternativas (`POST /questions`, `PUT /questions`, `DELETE /questions` retornam 404 Not Found).
- A gestão de questões é realizada pelo serviço interno `internal/usecase/question/service.go`.
- Ao criar um simulado (`POST /api/v1/exams`), o sistema seleciona automaticamente as questões no banco através do repositório/serviço interno de acordo com as áreas solicitadas ou de forma aleatória, vinculando-as ao exame através de `QuestionsOnExams`.
- Para carregar questões no banco de dados, utiliza-se a inicialização automática do sistema ou o comando CLI dedicado `cmd/seeder/main.go`.

---

## 🧪 4. Testes Automatizados e de Integração Separados

Para garantir visualização limpa e manutenção independente, os testes foram alocados em `tests/`, completamente apartados de `cmd/` e `internal/`:

| Diretório | Finalidade |
| :--- | :--- |
| `tests/mocks/` | Mocks em memória das interfaces de repositório (`UserRepository`, `QuestionRepository`, `ExamRepository`, `HistoryRepository`). |
| `tests/unit/` | Testes unitários de regras de negócio: validação de credenciais, hashing, isolamento de simulados, cálculo de pontuação, proibição de responder exame finalizado. |
| `tests/integration/` | Testes de integração HTTP cobrindo o fluxo completo de registro, login, barreira 401 sem token, barreira 403 para usuários cruzados, respostas e verificação de que `/questions` retorna 404. |

### Executando os testes:
```bash
# Executar todos os testes
make test

# Executar apenas testes unitários
make test-unit

# Executar apenas testes de integração
make test-integration
```

---

## 🗄️ 5. Controle de Versão do Esquema de Banco de Dados (Migrations)

O projeto possui um sistema próprio de **controle de versão incremental do esquema** implementado em `internal/infrastructure/database/migrator.go`, garantindo que alterações estruturais sejam aplicadas em ordem determinística e apenas uma vez:

- **Tabela de Auditoria (`schema_migrations`)**: Registra cada arquivo de migração executado e seu timestamp (`version`, `applied_at`).
- **Arquivos Versionados (`migrations/*.sql`)**:
  - `000001_init_schema.sql`: Definição inicial de tabelas (`users`, `questions`, `question_topics`, `question_images`, `alternatives`, `exams`, `exam_areas`, `questions_on_exams`, `histories`) e índices estruturais.
  - `000002_add_analytics_indexes.sql`: Índices compostos específicos para otimização de consultas analíticas e relatórios de performance (`idx_qoe_got_right`, `idx_histories_finished`, `idx_qoe_updated_at`, `idx_exams_user_date`).
- **Execução Automática e Transacional**:
  - Na inicialização da API (`cmd/api/main.go`), o migrador lê os arquivos embutidos via `embed.FS`, identifica quais versões ainda não foram aplicadas e as executa individualmente dentro de transações atômicas (`BEGIN ... COMMIT`).
  - Pode ser disparado manualmente via CLI com `make migrate` (`cmd/migrate/main.go`).

---

## 🚀 6. Como Executar o Projeto

### Pré-requisitos
- [Go 1.22+](https://go.dev/) instalado (ou Docker)
- PostgreSQL (ou usar Docker Compose)

### Opção A: Executando com Docker Compose (Recomendado)
```bash
# Iniciar PostgreSQL e a API automaticamente
make docker-up
```
A API estará disponível em `http://localhost:8080`.

### Opção B: Executando Localmente
1. Configure as variáveis de ambiente em um arquivo `.env` (baseie-se no `.env.example`):
```bash
cp .env.example .env
```

2. Execute as migrações e o servidor:
```bash
make run
```

3. (Opcional) Executar o seeder de questões:
```bash
make seed
```

---

## 📡 7. Endpoints da API

### Autenticação (Pública)

#### `POST /api/v1/auth/register`
Cria uma nova conta de usuário.
```json
{
  "name": "Maria Silva",
  "email": "maria@example.com",
  "password": "senha_segura_123"
}
```

#### `POST /api/v1/auth/login`
Autentica o usuário e retorna o token JWT.
```json
{
  "email": "maria@example.com",
  "password": "senha_segura_123"
}
```

---

### Perfil e Simulados (Protegidos - Header `Authorization: Bearer <TOKEN>`)

#### `GET /api/v1/auth/me`
Retorna dados do usuário autenticado.

#### `POST /api/v1/exams`
Cria um novo simulado para o usuário autenticado. As questões são selecionadas internamente.
```json
{
  "title": "Simulado de Humanas e Exatas",
  "areas": ["Ciências Humanas", "Matemática"],
  "question_count": 10
}
```

#### `GET /api/v1/exams`
Lista os simulados pertencentes exclusivamente ao usuário autenticado.

#### `GET /api/v1/exams/{id}`
Obtém os detalhes de um simulado do usuário (questões, alternativas, respostas salvas). Se o simulado não estiver finalizado, a resposta correta não é exibida.

#### `POST /api/v1/exams/{id}/submit`
Envia um lote de questões respondidas na modalidade de simulado:
```json
{
  "answers":[
    "question":{
      "question_id": "uuid-da-questao",
      "selected_alternative_id": "uuid-da-alternativa"
    }
  ]
}
```

#### `POST /api/v1/exams/{id}/answers`
Registra a resposta do usuário a uma questão na modalidade sprint:
```json
{
  "question_id": "uuid-da-questao",
  "selected_alternative_id": "uuid-da-alternativa"
}
```

#### `POST /api/v1/exams/{id}/finish`
Finaliza um simulado na modalidade sprint, calcula a nota (`score`), gera o registro em `History` e disponibiliza o gabarito.
```json
{
  "success": true,
  "data": {
    "exam_id": "uuid-do-simulado",
    "total_questions": 10,
    "correct_answers": 8,
    "score": 80,
    "finished_at": "2026-10-22T17:30:00Z"
  }
}
```

#### `DELETE /api/v1/exams/{id}`
Exclui um simulado do usuário.

---

### Histórico (Protegido - Header `Authorization: Bearer <TOKEN>`)

#### `GET /api/v1/histories`
Lista o histórico de simulados finalizados e notas do usuário autenticado.

#### `GET /api/v1/histories/{id}`
Detalha um registro de histórico do usuário.

---

### Análise de Performance & Métricas (Protegido - Header `Authorization: Bearer <TOKEN>`)

Todos os endpoints analíticos executam agregações otimizadas diretamente no banco PostgreSQL (`COUNT`, `SUM`, `AVG`, `FILTER (WHERE got_right = true)`), tratam divisões por zero com `NULLIF`, e isolam os dados exclusivamente para o usuário autenticado.

#### `GET /api/v1/analytics/overview`
Retorna a visão geral e resumo executivo:
- **Total de Questões Respondidas**
- **Taxa Geral de Acerto (%)**
- **Total de Simulados Concluídos**
- **Média Geral de Nota em Simulados**
- **Desempenho por Área de Conhecimento** (área, respondidas, acertos, %)
- **Ritmo de Estudo Recente** (questões respondidas por dia nos últimos 30 dias)

#### `GET /api/v1/analytics/by-discipline`
Retorna o desempenho detalhado por disciplina com volume e % de acerto.
- Query params opcionais:
  - `start_date` (ex: `2026-01-01` ou ISO 8601)
  - `end_date` (ex: `2026-12-31`)
  - `area` (ex: `Ciências da Natureza`)

#### `GET /api/v1/analytics/topics-ranking`
Identifica os pontos fortes e pontos fracos do usuário:
- **Top 5 Tópicos de Maior Domínio:** maior taxa de acerto com mínimo de 10 questões (configurável via `min_questions`).
- **Top 5 Tópicos Críticos/Pontos Fracos:** tópicos com menor taxa de acerto.
- Query params opcionais:
  - `min_questions` (padrão: `10`)

#### `GET /api/v1/analytics/evolution`
Retorna a série temporal da evolução do aluno nos últimos meses:
- Query params opcionais:
  - `period`: `weekly` (semanal, ex: `2026-W42`) ou `monthly` (mensal, ex: `2026-10`) - padrão: `weekly`.
  - `months`: número de meses de histórico (padrão: `6`).
- Retorna os pontos temporais e o ritmo diário nos últimos 30 dias (adequado para heatmaps e gráficos de barras).

#### `GET /api/v1/analytics/exams-history`
Retorna a lista cronológica de todos os simulados finalizados:
- Título, data, nota (`score`), total de questões e acertos.
- `overall_average_score`: Média de nota histórica de todos os simulados.
- `recent_average_score`: Média de nota dos simulados mais recentes.
- `score_evolution_percentage`: Evolução percentual calculada entre os simulados anteriores e recentes.

