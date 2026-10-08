# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Qué es

Aplicación web de lista de tareas en Go con **Buffalo** (v1.1) y **Pop** sobre **PostgreSQL**. Renderiza HTML en servidor con plantillas Plush; no hay API JSON salvo `/healthz`. Los textos de la interfaz, los mensajes de validación y los commits están en español.

## Comandos

El entorno previsto es el dev container (`.devcontainer/`), que instala Go, el CLI de `buffalo` y levanta Postgres con `docker compose up -d db`. Credenciales `postgres/postgres` en `127.0.0.1:5432` (sobrescribible con `DB_HOST`).

```sh
docker compose up -d db              # Postgres (crea solo todo_development)
go run ./cmd/app                     # arranca en :3000 y aplica las migraciones pendientes
buffalo dev                          # igual, con recarga en caliente (.buffalo.dev.yml)
docker compose up --build            # app + db en contenedores
```

Tests (requieren Postgres y la base `todo_test` migrada):

```sh
go install github.com/gobuffalo/pop/v6/soda@v6.4.1
GO_ENV=test soda create -e test      # solo la primera vez
GO_ENV=test soda migrate -e test
GO_ENV=test go test -p 1 ./...
GO_ENV=test go test -p 1 ./actions -run Test_ActionSuite -testify.m Test_TasksCreate_Valid   # un solo test
```

- **`GO_ENV=test` es obligatorio**: sin él, `models.DB` se conecta a `todo_development` y las suites truncan esa base.
- **`-p 1` es obligatorio**: `actions` y `models` comparten `todo_test` y truncan las tablas entre tests, así que en paralelo se pisan.

Lo que comprueba la CI (`.github/workflows/ci.yaml`) en cada PR: `gofmt -l .` sin salida, `go vet ./...`, los tests, `go build ./...` y `docker build`.

## Arquitectura

- **Punto de entrada** `cmd/app/main.go`: ejecuta `models.Migrate()` y luego `actions.App().Serve()`. Las migraciones van embebidas en el binario (`migrations/embed.go`), así que en desarrollo y producción no hace falta `soda`; solo se usa en la CI y en local para la base de test.
- **`models`**: su `init()` abre `models.DB` según `GO_ENV` usando `database.yml`, así que cualquier paquete que importe `models` necesita una configuración de BD válida. Las validaciones viven en `Validate()` del modelo y se disparan con `tx.ValidateAndCreate`/`ValidateAndUpdate`.
- **`actions`**: rutas y middleware en `app.go` (singleton con `sync.Once`). `popmw.Transaction` envuelve cada petición en una transacción: los handlers usan `c.Value("tx").(*pop.Connection)`, nunca `models.DB` directamente. Si el handler devuelve error se hace rollback.
- **Formularios HTML**: PUT y DELETE se envían como POST con el campo oculto `_method`. Como el middleware CSRF está activo, todo formulario necesita también `authenticity_token`. Ver `templates/tasks/index.plush.html`.
- **Patrón de los handlers**: si la validación falla, se vuelve a renderizar `index` con 422 y `errors`. Si va bien, se añade un flash y se redirige con 303 a `/`. Un id inexistente o que no es un UUID da 404.
- **Plantillas y estáticos** van embebidos con `go:embed` (`templates/embed.go`, `public/embed.go`): tras cambiarlos hay que recompilar. El patrón de `templates` es `* */*`, así que solo cubre un nivel de subcarpetas.
- **Migraciones**: Fizz para el esquema y SQL plano para los datos. `20260901000001_seed_tasks` inserta tres tareas de ejemplo con `ON CONFLICT DO NOTHING`, así que también llegan a producción.
- **Tests**: suites de `gobuffalo/suite` (testify). Los métodos son `Test_*` sobre `ActionSuite`/`ModelSuite`. Los datos de prueba se cargan con `LoadFixture("<scenario>")` desde `fixtures/*.toml`.
