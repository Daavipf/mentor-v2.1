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
│   └── seeder/
│       └── main.go              # CLI interno para povoamento de questões e alternativas
├── internal/
│   ├── domain/                  # Enterprise Business Rules (independente de frameworks externos)
│   │   ├── entity/              # Entidades: User, Question, Alternative, Exam, History, etc.
│   │   ├── repository/          # Interfaces (Ports) dos repositórios
│   │   └── errors.go            # Erros de domínio (ErrNotFound, ErrForbidden, etc.)
│   ├── usecase/                 # Application Business Rules (casos de uso)
│   │   ├── auth/                # Registro, login e perfil do usuário
│   │   ├── exam/                # Criação, listagem, resolução e finalização de simulados
│   │   ├── history/             # Histórico de notas e tentativas
│   │   └── question/            # Serviço interno de questões/alternativas (SEM exposição HTTP)
│   ├── infrastructure/          # Frameworks, Drivers & Adapters
│   │   ├── config/              # Variáveis de ambiente e configuração
│   │   ├── database/            # Conexão PostgreSQL (pgx) e executor de migrações embutidas
│   │   ├── repository/postgres/ # Implementações concretas dos repositórios em PostgreSQL
│   │   └── security/            # Hashing seguro com Bcrypt e emissão/validação de tokens JWT
│   └── interface/http/          # Interface Adapters (HTTP)
│       ├── handler/             # Controllers HTTP (AuthHandler, ExamHandler, HistoryHandler)
│       ├── middleware/          # Autenticação JWT, Logging estruturado, Recuperação e CORS
│       ├── response/            # Padronização de respostas JSON e mapeador de erros
│       └── router.go            # Configuração das rotas Chi v5
├── migrations/                  # Scripts SQL DDL correspondentes ao schema.prisma
│   └── 000001_init_schema.sql
├── tests/                       # Testes automatizados e de integração SEPARADOS do código-fonte
│   ├── mocks/                   # Implementações em memória dos repositórios para testes
│   ├── unit/                    # Testes unitários dos casos de uso e regras de negócio
│   └── integration/             # Testes de integração end-to-end das rotas HTTP
├── Dockerfile                   # Build multi-etapas minimalista em Alpine
├── docker-compose.yml           # Orquestração do banco PostgreSQL + API
├── Makefile                     # Atalhos para compilação, testes e execução
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

## 🚀 5. Como Executar o Projeto

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

## 📡 6. Endpoints da API

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

#### `POST /api/v1/exams/{id}/answers`
Registra a resposta do usuário a uma questão do simulado:
```json
{
  "question_id": "uuid-da-questao",
  "selected_alternative_id": "uuid-da-alternativa"
}
```

#### `POST /api/v1/exams/{id}/finish`
Finaliza o simulado, calcula a nota (`score`), gera o registro em `History` e disponibiliza o gabarito.
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

