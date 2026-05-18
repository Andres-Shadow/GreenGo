# GreenGo

GreenGo es una libreria de Go para crear pipelines simples y reutilizables. Su caso principal es vigilar una rama de Git y desplegar el proyecto con Docker Compose cada vez que aparece un nuevo commit.

El proyecto tambien incluye un comando listo para usar: `greengo`.

## Instalacion

```bash
go get github.com/Andres-Shadow/GreenGo
```

Para compilar el comando:

```bash
go build -o greengo ./cmd/greengo
```

## Uso como CLI

```bash
./greengo \
  -repo https://github.com/usuario/proyecto.git \
  -branch main \
  -workspace ./runtime/proyecto \
  -interval 30s \
  -initial-run
```

Por defecto GreenGo ejecuta:

```bash
docker compose up --build -d
```

Si necesitas Docker Compose v1:

```bash
./greengo -repo https://github.com/usuario/proyecto.git -compose-command docker-compose
```

Tambien puedes configurar el comando con variables de entorno:

```bash
GREENGO_REPO=https://github.com/usuario/proyecto.git
GREENGO_BRANCH=main
GREENGO_WORKSPACE=./runtime/proyecto
GREENGO_INTERVAL=30s
GREENGO_COMPOSE_FILES=docker-compose.yml,docker-compose.prod.yml
GREENGO_COMPOSE_COMMAND=docker
GREENGO_INITIAL_RUN=true
```

## Uso como libreria

### Pipeline generico

```go
package main

import (
	"context"
	"log"
	"os"

	greengo "github.com/Andres-Shadow/GreenGo"
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)

	pipeline := greengo.NewPipeline("build-and-test", greengo.WithLogger(logger))
	pipeline.AddCommand("download deps", greengo.Command{Name: "go", Args: []string{"mod", "download"}})
	pipeline.AddCommand("run tests", greengo.Command{Name: "go", Args: []string{"test", "./..."}})

	if err := pipeline.Run(context.Background(), greengo.RunContext{}); err != nil {
		logger.Fatal(err)
	}
}
```

### Pipeline de despliegue con Docker Compose

```go
package main

import (
	"context"
	"log"
	"os"
	"time"

	greengo "github.com/Andres-Shadow/GreenGo"
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)

	pipeline, err := greengo.NewDockerComposePipeline(greengo.DeployConfig{
		RepoURL:   "https://github.com/usuario/proyecto.git",
		Branch:    "main",
		Workspace: "./runtime/proyecto",
		Build:     true,
		Detach:    true,
	}, greengo.WithLogger(logger))
	if err != nil {
		logger.Fatal(err)
	}

	err = greengo.Watch(context.Background(), greengo.WatchConfig{
		RepoURL:    "https://github.com/usuario/proyecto.git",
		Branch:     "main",
		Interval:   30 * time.Second,
		Workspace:  "./runtime/proyecto",
		Pipeline:   pipeline,
		Logger:     logger,
		InitialRun: true,
	})
	if err != nil {
		logger.Fatal(err)
	}
}
```

## API principal

- `NewPipeline`: crea un pipeline reutilizable.
- `AddStage`: agrega una etapa con logica Go personalizada.
- `AddCommand`: agrega una etapa basada en comandos del sistema.
- `Run`: ejecuta las etapas en orden y se detiene ante el primer error.
- `LatestCommit`: consulta el ultimo commit remoto de una rama.
- `Watch`: vigila una rama y ejecuta un pipeline cuando cambia el commit.
- `NewDockerComposePipeline`: crea el flujo Git clone + Docker Compose.

## Flujo de despliegue incluido

`NewDockerComposePipeline` ejecuta estas etapas:

1. Limpia y crea el directorio de trabajo.
2. Clona la rama configurada con `git clone --branch <branch> --single-branch`.
3. Ejecuta Docker Compose en el proyecto clonado.

Este flujo encaja con proyectos universitarios o personales donde cada aplicacion ya trae su `docker-compose.yml` y solo necesitas redeplegar al detectar cambios en Git.

## Desarrollo

```bash
go test ./...
go vet ./...
```

## Creditos

La idea inicial se inspiro en `python-deployment-script-sample` de [acoronadoc](https://github.com/acoronadoc). GreenGo ahora separa esa premisa en una libreria Go y un CLI reutilizable para despliegues con Docker Compose.
