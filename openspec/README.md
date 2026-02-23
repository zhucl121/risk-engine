# OpenSpec — RiskEngine

This directory implements a spec-driven development workflow for the riskengine project,
inspired by [OpenSpec](https://github.com/Fission-AI/OpenSpec) and [Superpowers](https://supercode.sh).

## Philosophy

> Agree before you build. Every non-trivial change goes through specs before a single line of code is written.

## Structure

```
openspec/
├── specs/          ← Living master specification (source of truth)
│   ├── engine.md
│   ├── rule-engine.md
│   ├── feature-service.md
│   ├── model-engine.md
│   ├── list-service.md
│   ├── orchestrator.md
│   └── api.md
└── changes/
    ├── <active-feature>/   ← Current work in progress
    │   ├── proposal.md
    │   ├── specs/
    │   ├── design.md
    │   └── tasks.md
    └── archive/            ← Completed changes
```

## How to use

```
/opsx:new <feature-name>   → create change folder
/opsx:ff                   → fast-forward: generate all planning docs
/opsx:review               → architect review of specs/design
/opsx:apply                → sequential implementation of tasks
/opsx:archive              → archive completed change
```

Full workflow guide: `.cursor/skills/riskengine-workflow/SKILL.md`  
Role definitions + spec templates: `.cursor/skills/riskengine-workflow/`
