# Paterna

Docker observability and container management platform.

## Instalação

### macOS / Linux (curl)

```sh
curl -fsSL https://raw.githubusercontent.com/hugaojanuario/Paterna/main/install.sh | sh
```

O script detecta seu SO/arquitetura, baixa o binário da última release e
instala em `/usr/local/bin/paterna` (ou `~/.local/bin/paterna` se não houver
permissão de sudo).

### Build local

```sh
make build      # gera ./paterna
make install    # copia para ~/.local/bin/paterna
```

## Uso

```sh
paterna --help     # lista comandos
paterna version    # mostra versão / commit / data de build
paterna            # abre o TUI
```

## Desenvolvimento

```sh
make run           # roda direto via go run
make tidy          # go mod tidy
make release-dry   # testa goreleaser sem publicar
```

Para publicar uma release nova: crie uma tag `vX.Y.Z` e dê push. O workflow
`.github/workflows/release.yml` roda goreleaser e publica os binários.

```sh
git tag v0.2.0
git push origin v0.2.0
```
