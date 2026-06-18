
<p align="center">
  <img src="docs/images/logo-paterna.png" width="500">
</p>


Docker observability and container management platform — CLI/TUI escrito em Go.

Paterna é uma ferramenta de terminal para monitorar e gerenciar containers Docker. Roda um daemon em background no servidor que coleta métricas, dispara alertas e expõe uma API interna via Unix socket. A interface é um TUI interativo construído com Bubble Tea.

---

## Instalação

### macOS / Linux (curl)

```sh
curl -fsSL https://raw.githubusercontent.com/hugaojanuario/Paterna/main/install.sh | sh
```

O script detecta o sistema operacional e a arquitetura, baixa o binário da última release e instala em `/usr/local/bin/paterna`. Se não houver permissão de `sudo`, instala em `~/.local/bin/paterna`.

### Build local

```sh
make build      # compila e gera ./paterna
make install    # copia para ~/.local/bin/paterna
```

**Pré-requisitos:** Go 1.23+, Docker rodando localmente.

---

## Uso

```sh
paterna               # abre o TUI interativo
paterna init          # primeira execução: cria admin e sobe o daemon
paterna start         # sobe o daemon (docker start paterna-daemon)
paterna stop          # para o daemon (docker stop paterna-daemon)
paterna reload        # reinicia o daemon
paterna status        # mostra se o daemon está rodando
paterna logs          # mostra logs do daemon
paterna version       # versão, commit e data do build
paterna --help        # lista todos os comandos
```

---

## Arquitetura

```
usuário no terminal
      │
   CLI/TUI (paterna)
      │  Unix socket (/var/run/paterna.sock)
      ▼
   Daemon (container Docker em background)
      │
  ┌───────────────┬───────────────┬───────────────┐
  │ container-svc │  metrics-svc  │   alert-svc   │
  └───────────────┴───────────────┴───────────────┘
         │               │               │
      Docker API      Prometheus       Telegram
         │               │
      PostgreSQL      PostgreSQL
```

- **CLI** — binário `paterna` instalado no servidor. Abre o TUI e envia comandos ao daemon via Unix socket.
- **Daemon** — roda como container Docker. Contém toda a lógica de negócio e expõe uma API HTTP interna no socket.
- **Unix socket** — `/var/run/paterna.sock`. Comunicação local entre CLI e daemon, sem expor porta na rede.

---

## Stack

| Camada | Tecnologia |
|--------|-----------|
| Linguagem | Go 1.23+ |
| CLI | Cobra |
| TUI | Bubble Tea + Lip Gloss + Bubbles |
| Docker | Docker SDK for Go |
| Banco | PostgreSQL + golang-migrate |
| Métricas | Prometheus |
| Alertas | Telegram |
| Infra | Docker, GitHub Actions, Terraform, Kubernetes |
| Release | GoReleaser |

---

## Estrutura do Projeto

```
paterna/
├── cmd/
│   └── cli/
│       └── main.go              # entrypoint da CLI
├── internal/
│   ├── cli/                     # subcomandos Cobra (init, start, stop…)
│   ├── tui/
│   │   ├── app.go               # entrada do TUI
│   │   ├── models/              # telas: menu, containers, métricas, alertas…
│   │   └── styles/              # cores e estilos Lip Gloss
│   ├── daemon/                  # HTTP server no Unix socket + rotas
│   ├── container/               # handler, service, repository
│   ├── metrics/                 # coleta de CPU/memória, histórico
│   ├── alert/                   # regras, Telegram, histórico
│   └── shared/                  # auth JWT, config, client socket, database
├── db/
│   └── migrations/              # SQL migrations numeradas
├── pkg/
├── www/
├── .github/
│   └── workflows/
│       └── release.yml          # goreleaser automático por tag
├── .goreleaser.yaml
├── Makefile
└── go.mod
```

---

## Desenvolvimento

```sh
make run           # roda via go run (sem compilar)
make tidy          # go mod tidy
make clean         # remove binário e dist/
make release-dry   # testa goreleaser sem publicar
```

### Publicar uma release

Crie uma tag `vX.Y.Z` e dê push. O workflow `.github/workflows/release.yml` executa o GoReleaser e publica os binários automaticamente.

```sh
git tag v0.3.0
git push origin v0.3.0
```

---

## Variáveis de Ambiente

Copie `.env.example` para `.env` e preencha:

```sh
cp .env.example .env
```

| Variável | Descrição |
|----------|-----------|
| `DATABASE_URL` | URL de conexão PostgreSQL |
| `JWT_SECRET` | Segredo para assinar tokens JWT |
| `TELEGRAM_BOT_TOKEN` | Token do bot para alertas |
| `TELEGRAM_CHAT_ID` | Chat ID para receber alertas |

---

## Licença

[MIT](LICENSE)
