# 🔥 Hefesto

**El Forge del Ecosistema Techne**

> Configuración y forjado del entorno de desarrollo con IA — simplificado, potente, sin excusas.

---

## ¿Qué es Hefesto?

Hefesto es un sistema de configuración para entornos de desarrollo asistidos por IA. Proporciona un workflow **Spec-Driven Development (SDD)** simplificado, un agente mentor integrado, y persistencia de memoria跨-sesiones.

A diferencia de otras soluciones sobre-ingenierizadas, Hefesto se enfoca en lo esencial:
- **Menos configuración, más código real**
- **Persistencia simple sin mode branching**
- **Agente que enseña, no que solo ejecuta**

---

## Parte del Ecosistema Techne

Hefesto es uno de los pilares del ecosistema Techne:

| Proyecto | Rol |
|----------|-----|
| **Techne** | La matriz/fundación del ecosistema |
| **Apolo** | Agente IA para servidores VPS |
| **Artemisa** | Agente IA para servidores VPS |
| **Hefesto** | El forge — configura y forja el entorno de desarrollo |
| **Mneme** | Fork de Engram (memoria persistente, próximamente) |

---

## Características Principales

### 🔄 SDD Workflow Simplificado
6 fases claras vs las 9 de Gentleman.Dots:
```
init → explore → spec → tasks → apply → verify
```

### 💾 Persistencia Simple
- **Un solo modo**: Engram (memoria persistente)
- Sin `openspec`, sin `hybrid`, sin `none`
- Sin branching de configuración por modo

### 🤖 Agente Hefesto
- Personalidad mentor — enseña fundamentos, no shortcuts
- Helpful-first pero sin tolerar mediocridad
- Responde en el idioma del usuario

### 🌐 Remote-Exec de Primera Clase
- SSH/VPS como ciudadanos de primera
- Sin bloqueos por operaciones remotas
- Delegación clara al sub-agente `remote-exec`

### 🔥 Tema Fuego/Forge
- Identidad visual coherente
- Emojis y metáforas de herrería/fuego

### 📦 Skill System Extensible
- Sistema modular de skills
- Registry por proyecto
- Fácil adición de nuevas capacidades

---

## Comparación con Gentleman.Dots

| Aspecto | Gentleman.Dots | Hefesto |
|---------|----------------|---------|
| Fases SDD | 9 | 6 |
| Modos persistencia | 4 (engram, openspec, hybrid, none) | 1 (engram) |
| Config branching | Sí (por modo) | No |
| SSH handling | Bloquea main loop | Delegación async |
| Skills incluidas | 18+ | Esencial (creciendo) |
| Complejidad config | ~100% | ~62% |

---

## Estructura del Proyecto

```
Hefesto/
├── README.md
├── LICENSE
├── .gitignore
├── docs/
│   └── SDD.md                 # Documentación del workflow
└── HefestoOpenCode/
    ├── AGENTS.md              # Personalidad y reglas del agente
    ├── GEMINI.md              # Instrucciones específicas de Gemini
    ├── commands/              # Comandos slash personalizados
    └── skills/                # Skills modulares
        ├── sdd-init/
        ├── sdd-explore/
        ├── sdd-propose/
        ├── sdd-spec/
        ├── sdd-design/
        ├── sdd-tasks/
        ├── sdd-apply/
        ├── sdd-verify/
        └── sdd-archive/
```

---

## Instalación

### Próximamente
- Instalador TUI interactivo
- Homebrew formula (`brew install hefesto`)

### Instalación Manual

```bash
# Clonar el repositorio
git clone https://github.com/Edcko/Hefesto.git
cd Hefesto

# Copiar configuración a OpenCode
cp -r HefestoOpenCode ~/.config/opencode/

# Verificar instalación
# Reinicia tu sesión de OpenCode/Gemini CLI
```

---

## Quick Start: SDD Workflow

```bash
# 1. Inicializar SDD en tu proyecto
/sdd-init

# 2. Explorar una idea antes de comprometerte
/sdd-explore mi-feature

# 3. Crear propuesta de cambio
/sdd-new agregar-autenticacion

# 4. O ir rápido con todas las fases de planeación
/sdd-ff agregar-autenticacion

# 5. Implementar
/sdd-apply agregar-autenticacion

# 6. Verificar
/sdd-verify agregar-autenticacion
```

### Diagrama de Fases

```
┌─────────┐   ┌──────────┐   ┌────────┐   ┌────────┐   ┌────────┐   ┌────────┐
│  INIT   │ → │ EXPLORE  │ → │  SPEC  │ → │ TASKS  │ → │ APPLY  │ → │ VERIFY │
└─────────┘   └──────────┘   └────────┘   └────────┘   └────────┘   └────────┘
     │              │              │            │            │            │
     ▼              ▼              ▼            ▼            ▼            ▼
  Contexto      Investigar    Requisitos   Checklist    Código      Validar
  del stack     y clarificar  y escenarios  de tasks    real        specs
```

---

## Contribuir

1. Fork el repositorio
2. Crea una rama: `git checkout -b feature/mi-mejora`
3. Commit: `git commit -m "feat: descripción clara"`
4. Push: `git push origin feature/mi-mejora`
5. Abre un Pull Request

**Reglas de oro:**
- Commits convencionales (`feat:`, `fix:`, `docs:`, `refactor:`)
- Sin atribución de IA en commits
- Código limpio sobre código rápido

---

## Licencia

MIT License © 2026 Edcko

---

> 🔥 *"El buen herrero no culpa a su martillo. Forja con lo que tiene."*
