# Paterna

> Observabilidade e gerenciamento de containers Docker — TUI, CLI e Web em um único binário.

Paterna é uma plataforma para monitorar, inspecionar e controlar containers
Docker. Oferece três interfaces que falam com o mesmo núcleo:

- **TUI** interativa no terminal, no estilo `k9s`/`lazydocker`
- **CLI** para automação e scripts
- **API REST** consumida por uma interface web

Escrito em Go, sem dependências externas pesadas — SQLite local para
persistência, Docker SDK oficial para comunicação com o engine.

---

## Instalação

### Linux e macOS — uma linha

```bash
curl -fsSL https://raw.githubusercontent.com/hugaojanuario/Paterna/main/install.sh | sh
```

O script detecta seu sistema (Linux ou macOS, amd64 ou arm64), baixa o binário
mais recente do GitHub Releases e instala em `/usr/local/bin/paterna`.

### A partir do código-fonte

```bash
git clone https://github.com/hugaojanuario/Paterna.git
cd Paterna
go build -o paterna ./cmd/cli
sudo mv paterna /usr/local/bin/
```

### Via `go install`

```bash
go install github.com/hugaojanuario/Paterna/cmd/cli@latest
```

> O binário será criado como `cli` em `$GOPATH/bin`. Renomeie ou crie um alias
> para `paterna`.

---

## Primeira execução

```bash
paterna init
```

O comando solicita email e senha do administrador, cria o banco SQLite local
e prepara o ambiente. Execute apenas uma vez.

```
Paterna — Configuração inicial

Email: admin@exemplo.com
Senha:
Confirme a senha:

Conta criada com sucesso.

Próximos passos:
  paterna         abre o painel TUI
  paterna serve   sobe a API HTTP em :8080
```

---

## Uso

### Painel TUI

```bash
paterna
```

Solicita login com as credenciais cadastradas. Após autenticação, abre o
painel interativo de containers.

| Tecla | Ação |
|---|---|
| `↑` `↓` | Navegar na lista |
| `s` | Iniciar container |
| `x` | Parar container |
| `r` | Reiniciar container |
| `u` | Atualizar lista |
| `q` ou `Ctrl+C` | Sair |
| `tab` (login) | Alternar entre campos |
| `enter` (login) | Entrar |
| `esc` (login) | Sair |

### API HTTP

```bash
paterna serve
```

Sobe o servidor HTTP em `:8080`. Override de porta:

```bash
paterna serve --port 9000
# ou
API_PORT=9000 paterna serve
```

---

## Stack

| Camada | Tecnologia |
|---|---|
| Linguagem | Go 1.24+ |
| CLI | Cobra |
| TUI | Bubble Tea, Lip Gloss, Bubbles |
| Web | Pierrot (framework próprio) |
| API | `net/http` (stdlib) |
| Persistência | SQLite (`modernc.org/sqlite`, pure Go, sem CGO) |
| Hash | bcrypt |
| Docker | Docker Engine SDK for Go |
| Release | GoReleaser + GitHub Actions |

---

## Estrutura

```
.
├── cmd/cli/                binário paterna
├── internal/
│   ├── commands/           subcomandos Cobra (init, serve, auth)
│   ├── tui/                telas Bubble Tea (login, containers)
│   ├── container/          service Docker
│   ├── repository/         acesso ao SQLite (users)
│   └── api/                handlers HTTP, middleware, router
├── pkg/
│   ├── docker/             cliente Docker singleton
│   ├── database/           conexão SQLite
│   ├── session/            store de sessão em memória
│   ├── bcrypt/             wrapper de hash
│   ├── dotenv/             carregamento de variáveis
│   └── errors/             erros compartilhados
├── db/migrations/          migrations SQL
├── www/                    aplicação web (Pierrot)
├── install.sh              script de instalação
├── .goreleaser.yml         configuração de release
└── .github/workflows/      CI/CD
```

---

## API

### Autenticação

Login com credenciais cadastradas no `paterna init`:

```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@exemplo.com", "password": "suasenha"}'
```

Resposta:

```json
{ "token": "uuid-da-sessao", "email": "admin@exemplo.com" }
```

Todas as rotas seguintes exigem o header:

```
Authorization: Bearer <token>
```

### Endpoints

| Método | Rota | Descrição |
|---|---|---|
| `POST` | `/login` | Autentica e retorna token de sessão |
| `POST` | `/logout` | Invalida o token atual |
| `GET` | `/containers` | Lista containers |
| `POST` | `/containers/{id}/start` | Inicia container |
| `POST` | `/containers/{id}/stop` | Para container |
| `POST` | `/containers/{id}/restart` | Reinicia container |
| `GET` | `/containers/{id}/logs` | Últimas 50 linhas de log |
| `GET` | `/containers/{id}/stats` | CPU% e memória |
| `GET` | `/containers/{id}/inspect` | Estado detalhado |

### Exemplos

```bash
# listar
curl http://localhost:8080/containers \
  -H "Authorization: Bearer <token>"

# parar
curl -X POST http://localhost:8080/containers/abc123/stop \
  -H "Authorization: Bearer <token>"
```

---

## Comandos

```
paterna init              cria conta de administrador (primeira execução)
paterna                   abre o painel TUI (pede login)
paterna serve             sobe a API HTTP (porta padrão 8080)
paterna auth --register   adiciona um novo usuário ao banco
paterna auth --login      atualiza token de acesso local
```

---

## Banco de dados

SQLite local em `pkg/database/paterna.db` (durante desenvolvimento) ou no
diretório de execução em produção.

Inspecionar:

```bash
sqlite3 pkg/database/paterna.db
.tables
SELECT id, email, created_at FROM users;
.quit
```

Resetar:

```bash
rm pkg/database/paterna.db
paterna init
```

---

## Requisitos

- **Docker** rodando localmente (Docker Desktop, colima, ou daemon Linux)
- Permissão para acessar o socket Docker (`/var/run/docker.sock` no Linux/macOS)

---

## Roadmap

- [x] CLI com Cobra
- [x] TUI com Bubble Tea (login + lista + start/stop/restart)
- [x] SQLite + bcrypt para autenticação local
- [x] API REST com sessões
- [x] Aplicação web integrada
- [x] Binário multi-platform via GoReleaser
- [x] Script de instalação por curl
- [ ] Tela de logs em tempo real na TUI
- [ ] Tela de métricas (CPU, memória, histórico)
- [ ] Sistema de alertas com regras configuráveis
- [ ] Notificações via Telegram
- [ ] Daemon rodando como container Docker

---

## Releasing

Tag e push:

```bash
git tag v0.1.0
git push origin v0.1.0
```

O workflow do GitHub Actions roda o GoReleaser automaticamente, gera binários
para Linux/macOS/Windows × amd64/arm64 e publica em GitHub Releases.

---

## Licença

MIT — veja [LICENSE](LICENSE).

---

## Autor

Hugo Januario — [@hugaojanuario](https://github.com/hugaojanuario)
