[🇬🇧 English](README.md)

```
    🔥
   ╱│╲
  ╱ │ ╲
 ╱  │  ╲        HEFESTO
╱___▼___╲       Forja de Entornos de Desarrollo con IA
 ║███████║
 ║███████║       Forja tu entorno perfecto de desarrollo con IA
 ╰═══════╯
```

[![Versión de Go](https://img.shields.io/badge/Go-1.26.1-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Licencia](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/Edcko/Hefesto?include_prereleases)](https://github.com/Edcko/Hefesto/releases)

---

## ¿Qué es Hefesto?

Hefesto es un instalador opinado y gestor de configuración para OpenCode. Despliega un entorno completo de desarrollo con IA: agentes, skills, flujo de trabajo SDD, memoria persistente y más — en un solo comando.

Sin configuraciones complejas. Sin copiar archivos manualmente. Solo instala y empieza a construir con asistencia de IA que realmente entiende tu flujo de trabajo.

---

## Características

- **10 Comandos CLI**: `install`, `status`, `update`, `uninstall`, `rollback`, `doctor`, `config`, `list`, `version`, `completion`
- **26 Skills de IA**: Angular, React, Next.js, TypeScript, Tailwind, Zod, Django, .NET, Playwright, Pytest y más
- **Flujo SDD de 6 Fases**: `init → plan → spec → tasks → apply → verify`
- **10 Agentes**: Mentor principal, orquestador, fases SDD y ejecución remota
- **Memoria Persistente**: Integración con Engram para contexto entre sesiones
- **Plugin de Agentes en Segundo Plano**: Ejecución paralela de tareas sin bloquear
- **Instalador TUI Interactivo**: Seguimiento de progreso con Bubbletea
- **Distribución vía Homebrew**: `brew install edcko/tap/hefesto`
- **Multi-Plataforma**: Binarios para darwin/linux × arm64/amd64 + android-arm64
- **Tema Fuego/Forja**: Identidad visual cohesiva con paleta ámbar/cobre

---

## Inicio Rápido

```bash
# Instalar vía Homebrew (macOS/Linux)
brew install edcko/tap/hefesto

# O descargar el binario desde GitHub Releases
# https://github.com/Edcko/Hefesto/releases

# Usuarios de Android/Termux: descarguen el binario android-arm64
# directamente desde GitHub Releases

# Instalar la configuración de Hefesto
hefesto install

# Verificar el estado de la instalación
hefesto doctor

# Empezar a usar OpenCode
opencode
```

---

## Comandos CLI

| Comando | Descripción | Flags |
|---------|-------------|-------|
| `hefesto install` | Instalar archivos de configuración de Hefesto | `--yes`, `--dry-run` |
| `hefesto status` | Mostrar estado de la instalación | `--verbose` |
| `hefesto doctor` | Ejecutar diagnósticos completos de salud | — |
| `hefesto update` | Actualizar a la última configuración (no el binario) | `--yes`, `--dry-run` |
| `hefesto uninstall` | Eliminar la configuración de Hefesto | `--yes`, `--purge` |
| `hefesto rollback` | Restaurar un respaldo anterior | `--yes`, `--list` |
| `hefesto config show` | Mostrar rutas de configuración actuales | — |
| `hefesto config path` | Imprimir ruta del directorio de configuración | — |
| `hefesto list skills` | Listar todos los skills incluidos | `--json` |
| `hefesto list themes` | Listar temas disponibles | `--json` |
| `hefesto list backups` | Listar respaldos con timestamp | `--json` |
| `hefesto version` | Imprimir información de versión | — |

### Detalles de Comandos

**`hefesto install`**
- Detecta el directorio de configuración de OpenCode (`~/.config/opencode/`)
- Crea respaldos con timestamp de configuraciones existentes
- Despliega los archivos de configuración embebidos
- Configura skills, temas, plugins y comandos
- Flags: `--yes` (no interactivo), `--dry-run` (previsualizar cambios)

**`hefesto update`**
- Crea respaldo con timestamp de la configuración actual
- Superpone los últimos archivos de configuración embebidos (preserva personalizaciones donde sea posible)
- **Importante**: Esto solo actualiza los archivos de configuración, NO el binario de Hefesto
- Para actualizar el binario: `brew upgrade hefesto` o descargar desde [GitHub Releases](https://github.com/Edcko/Hefesto/releases)
- Flags: `--yes` (saltar confirmación), `--dry-run` (previsualizar cambios)

**`hefesto doctor`**
Ejecuta verificaciones completas de:
- Estructura del directorio de configuración
- Validez del archivo AGENTS.md
- Configuración de opencode.json
- Directorio y estructura de skills
- Directorio de plugins (engram, background-agents)
- Configuración de tema
- Configuración de personalidad
- Comandos personalizados

**`hefesto rollback`**
- Lista respaldos disponibles con timestamps
- Crea respaldo de seguridad antes de restaurar
- Flags: `--list` (mostrar respaldos), `--yes` (restaurar el más reciente sin preguntar)

**`hefesto status`**
- Muestra directorio de instalación
- Despliega versión instalada
- Lista cantidad de skills disponibles
- Reporta salud de la configuración
- Flag: `--verbose` para salida detallada

---

## Autocompletado de Shell

Hefesto soporta autocompletado para bash, zsh, fish y PowerShell:

```bash
# Bash
hefesto completion bash > ~/.config/hefesto/completion.bash
source ~/.config/hefesto/completion.bash

# Zsh
hefesto completion zsh > "${fpath[1]}/_hefesto"

# Fish
hefesto completion fish > ~/.config/fish/completions/hefesto.fish
```

---

## ¿Qué se Instala?

Hefesto despliega lo siguiente en `~/.config/opencode/`:

```
~/.config/opencode/
├── AGENTS.md              # Persona Hefesto + orquestador SDD + protocolo Engram
├── opencode.json          # 10 definiciones de agentes con límites de pasos
├── skills/                # 26 directorios de skills
│   ├── _shared/           # Patrones comunes, convenciones de persistencia
│   ├── ai-sdk-5/          # Patrones Vercel AI SDK 5
│   ├── angular/           # Arquitectura Angular 20+
│   ├── django-drf/        # Django REST Framework
│   ├── dotnet/            # .NET 9 / ASP.NET Core
│   ├── go-testing/        # Testing en Go + Bubbletea TUI
│   ├── nextjs-15/         # Next.js 15 App Router
│   ├── playwright/        # Testing E2E con Playwright
│   ├── pr-review/         # Workflow de review de PRs en GitHub
│   ├── pytest/            # Patrones de testing en Python
│   ├── react-19/          # React 19 con Compiler
│   ├── remote-exec/       # Ejecución remota vía SSH
│   ├── sdd-*/             # Skills de fases SDD (init, plan, spec, tasks, apply, verify)
│   ├── skill-creator/     # Creación de skills para agentes de IA
│   ├── skill-registry/    # Gestión del registry de skills del proyecto
│   ├── stream-deck/       # Presentaciones y slide decks
│   ├── tailwind-4/        # Patrones Tailwind CSS 4
│   ├── technical-review/  # Workflow de evaluación técnica
│   ├── typescript/        # Patrones estrictos de TypeScript
│   ├── zod-4/             # Validación de schemas Zod 4
│   └── zustand-5/         # Manejo de estado Zustand 5
├── plugins/
│   ├── engram.ts          # Integración de memoria persistente
│   └── background-agents.ts  # Ejecución paralela de agentes
├── commands/              # 5 slash commands de SDD
│   ├── sdd-init.md
│   ├── sdd-new.md
│   ├── sdd-ff.md
│   ├── sdd-apply.md
│   └── sdd-verify.md
├── themes/
│   └── hefesto.json       # Tema Fuego/Forja (ámbar/cobre)
└── personality/
    └── hefesto.md         # Definición de la persona Hefesto
```

---

## Skills (26 en Total)

### Flujo SDD
- **sdd-init** — Inicializa contexto SDD y configuración del proyecto
- **sdd-plan** — Explora el codebase y crea propuestas de cambio (explorar + proponer fusionados)
- **sdd-spec** — Escribe especificaciones detalladas a partir de propuestas
- **sdd-tasks** — Descompone specs y diseños en tareas de implementación
- **sdd-apply** — Implementa cambios de código a partir de definiciones de tareas
- **sdd-verify** — Valida la implementación contra las especificaciones

### Frameworks Frontend
- **angular** — Arquitectura Angular 20+ con Scope Rule, Screaming Architecture, componentes standalone, signals
- **react-19** — Patrones React 19 con React Compiler (no se necesita useMemo/useCallback)
- **nextjs-15** — Patrones Next.js 15 App Router (routing, Server Actions, data fetching)
- **tailwind-4** — Patrones y mejores prácticas de Tailwind CSS 4 (cn(), theme variables)
- **typescript** — Patrones estrictos y mejores prácticas de TypeScript (types, interfaces, generics)
- **zustand-5** — Patrones de manejo de estado Zustand 5

### Frameworks Backend
- **django-drf** — Patrones Django REST Framework (ViewSets, Serializers, Filters)
- **dotnet** — .NET 9 / ASP.NET Core con Minimal APIs, Clean Architecture, EF Core

### Testing
- **playwright** — Patrones de testing E2E con Playwright (Page Objects, selectores, workflow MCP)
- **pytest** — Patrones de testing en Python con Pytest (fixtures, mocking, markers)
- **go-testing** — Tests en Go y testing de TUI Bubbletea

### IA y SDK
- **ai-sdk-5** — Patrones Vercel AI SDK 5 (cambios rompientes desde v4)
- **zod-4** — Patrones de validación de schemas Zod 4 (cambios rompientes desde v3)

### Workflow y Review
- **pr-review** — Review de PRs e Issues de GitHub con análisis estructurado
- **technical-review** — Evaluación de ejercicios técnicos y entregas de candidatos
- **skill-creator** — Crear nuevos skills para agentes de IA siguiendo el spec Agent Skills
- **skill-registry** — Crear o actualizar el registry de skills del proyecto
- **stream-deck** — Crear slide decks de presentación para streams y cursos

### DevOps
- **remote-exec** — Ejecutar comandos en servidores remotos vía SSH

### Compartidos
- **_shared** — Patrones comunes de fases, convenciones de persistencia

---

## Flujo SDD

Spec-Driven Development es la capa de planeación estructurada para cambios sustanciales.

### 6 Fases

```
init → plan → spec → tasks → apply → verify
```

- **init** — Inicializa contexto SDD y detecta el stack del proyecto
- **plan** — Explora el codebase + crea propuesta de cambio (explorar + proponer fusionados)
- **spec** — Escribe requerimientos y escenarios detallados
- **tasks** — Descompone specs en checklist de implementación
- **apply** — Implementa cambios de código a partir de definiciones de tareas
- **verify** — Valida que la implementación coincida con las especificaciones

### Slash Commands

| Comando | Acción |
|---------|--------|
| `/sdd-init` | Inicializar SDD en tu proyecto |
| `/sdd-new <cambio>` | Crear nuevo cambio (ejecuta la fase plan) |
| `/sdd-ff <cambio>` | Avance rápido: plan → spec → tasks |
| `/sdd-apply <cambio>` | Implementar tareas |
| `/sdd-verify <cambio>` | Validar implementación |

### Persistencia

**Solo Engram.** Sin selección de modo. Los artefactos se almacenan en memoria persistente y sobreviven sesiones y compactaciones.

---

## Arquitectura

### Instalador
- **Lenguaje**: Go 1.26.1
- **Framework TUI**: Bubbletea (Charmbracelet)
- **Framework CLI**: Cobra
- **Embebido**: Archivos de configuración embebidos en el binario vía `go:embed`

### Distribución
- **Homebrew Tap**: `edcko/tap/hefesto`
- **GitHub Releases**: 5 binarios por plataforma (darwin/linux × arm64/amd64 + android-arm64)
- **Tamaño de Instalación**: ~15MB (incluye todas las configs, skills, temas)

### Configuración
- **Directorio Destino**: `~/.config/opencode/`
- **Estrategia de Respaldo**: Respaldos con timestamp antes de cada operación
- **Soporte de Rollback**: Restauración completa de respaldos con respaldo de seguridad

---

## Desarrollo

```bash
# Clonar el repositorio
git clone https://github.com/Edcko/Hefesto.git
cd Hefesto

# Construir el binario
cd cmd/hefesto
go build .

# Ejecutar localmente
./hefesto install --dry-run

# Probar en Docker (multi-plataforma)
cd ../..
./scripts/test.sh

# Correr tests unitarios
cd cmd/hefesto
go test ./...
```

### Estructura del Proyecto

```
Hefesto/
├── README.md
├── README.es.md
├── LICENSE
├── .gitignore
├── HefestoOpenCode/           # Configuración a desplegar
│   ├── AGENTS.md
│   ├── opencode.json
│   ├── skills/
│   ├── plugins/
│   ├── commands/
│   ├── themes/
│   └── personality/
├── cmd/hefesto/               # Binario instalador
│   ├── main.go                # Entry point del CLI
│   ├── internal/
│   │   ├── install/           # Lógica de instalación
│   │   ├── tui/               # TUI con Bubbletea
│   │   └── embed/config/      # Configs embebidos (vía go:embed)
│   └── go.mod
└── scripts/
    └── test.sh                # Runner de tests en Docker
```

---

## Ecosistema Techne

Hefesto es parte del ecosistema Techne:

| Proyecto | Rol |
|----------|-----|
| **Techne** | La fundación/matriz del ecosistema |
| **Hefesto** | La forja — configura y despliega entornos de desarrollo con IA |
| **Engram** | Memoria persistente para agentes de IA (integrado) |
| **OpenCode** | La plataforma de asistencia de código con IA |

---

## Contribuir

1. Haz fork del repositorio
2. Crea una rama de feature: `git checkout -b feature/mi-mejora`
3. Haz tus cambios siguiendo conventional commits
4. Commit: `git commit -m "feat: descripción clara"`
5. Push: `git push origin feature/mi-mejora`
6. Abre un Pull Request

**Lineamientos:**
- Usa conventional commits (`feat:`, `fix:`, `docs:`, `refactor:`)
- SIN atribución de IA en los commits (sin líneas "Co-Authored-By")
- Código limpio sobre código rápido
- Prueba tus cambios con `hefesto install --dry-run`

---

## Licencia

Licencia MIT © 2026 Edcko

Ver [LICENSE](LICENSE) para el texto completo.

---

> 🔥 *"Un buen herrero no le echa la culpa a su martillo. Forja con lo que tiene."*
