
<p align="center">
  <img src="docs/images/logo-paterna.png" width="500">
</p>


Monitor de sistema e containers Docker no terminal — TUI estilo [btop](https://github.com/aristocratos/btop), escrito em Go.

Paterna mostra CPU (por core), memória, swap, disco, rede e os processos mais pesados da máquina onde roda, além de um painel Docker com seus containers e stream de logs em tempo real. É 100% CLI: um binário, sem servidor, sem login, sem banco.

---

## Instalação

### macOS / Linux (curl)

```sh
curl -fsSL https://raw.githubusercontent.com/hugaojanuario/Paterna/main/install.sh | sh
```

O script detecta o sistema operacional e a arquitetura, baixa o binário da última release e instala em `/usr/local/bin/paterna`. Sem permissão de `sudo`, instala em `~/.local/bin/paterna`.

### Build local

```sh
make build      # compila e gera ./paterna
make install    # copia para ~/.local/bin/paterna
```

**Pré-requisitos:** Go 1.26+. Docker é opcional — sem ele, o painel Docker apenas avisa que está indisponível; o resto do monitor funciona normalmente.

---

## Uso

```sh
paterna           # abre o dashboard de monitoramento
paterna version   # versão, commit e data do build
paterna --help    # lista os comandos
```

### Teclas

**Dashboard**

| Tecla | Ação |
|-------|------|
| `d` / `enter` | Abre o gerenciador de containers Docker |
| `q` / `ctrl+c` | Sair |

**Containers** (gerenciador)

| Tecla | Ação |
|-------|------|
| `↑↓` / `jk` | Navegar |
| `enter` | Detalhes + logs em tempo real |
| `s` / `x` / `r` | start / stop / restart |
| `u` | Atualizar |
| `esc` | Voltar ao dashboard |

---

## Arquitetura

```
            paterna (binário único)
                    │
        ┌───────────┴───────────┐
        ▼                       ▼
  internal/system          internal/container
   (gopsutil)               (Docker SDK)
   CPU · MEM · DISK          listar · start/stop
   NET · PROC                stats · logs stream
        │                       │
        └──────────┬────────────┘
                   ▼
            internal/tui
       dashboard estilo btop (Bubble Tea)
```

Tudo é local: a TUI lê as métricas da máquina via gopsutil e fala com o Docker pelo socket local. Não há processo em background, porta de rede nem autenticação.

---

## Stack

| Camada | Tecnologia |
|--------|-----------|
| Linguagem | Go 1.26+ |
| CLI | Cobra |
| TUI | Bubble Tea + Lip Gloss + Bubbles |
| Métricas de sistema | gopsutil/v4 |
| Docker | Docker SDK for Go |
| Release | GoReleaser + GitHub Actions |

---

## Estrutura do Projeto

```
paterna/
├── cmd/
│   └── cli/main.go          # entrypoint
├── internal/
│   ├── commands/            # subcomandos Cobra (root, version)
│   ├── system/              # coleta de métricas via gopsutil
│   ├── container/           # serviço Docker (listar, stats, logs…)
│   ├── tui/
│   │   ├── dashboard.go     # tela principal estilo btop
│   │   ├── containers.go    # lista/ações de containers
│   │   ├── container_details.go  # detalhes + log stream
│   │   └── helpers.go       # helpers de render compartilhados
│   └── version/
├── pkg/
│   ├── docker/              # cliente Docker compartilhado
│   └── errors/              # erros sentinela
├── .github/workflows/release.yml
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
go test ./...      # smoke tests
```

### Publicar uma release

Crie uma tag `vX.Y.Z` e dê push. O workflow `.github/workflows/release.yml` executa o GoReleaser e publica os binários.

```sh
git tag v0.3.0
git push origin v0.3.0
```

---

## Licença

[MIT](LICENSE)
