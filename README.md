<div align="center">

# Notiflex Platform

### A GitAIOps public reference, built end to end with an AI agent

**English** | [도서 안내 (한국어)](https://github.com/gitaiops/_Book_GitAIOps)

</div>

---

## What this is

A B2B notification platform built on GKE by Claude Code, following the book
"AI 시대에 개발자가 알아야 할 인프라 구성 배포 with 클로드 코드" (ch2 to ch8, including the 8.3 CronJob), plus the generated onboarding guide from ch9.
Every step was recorded as it happened: see [`JOURNEY.md`](JOURNEY.md).

The point of this repository is not the app. It is the **recorded memory**:
plans, decisions, guardrails, and pinned values all live in Git, so both a new
engineer and a new AI agent can start from the same source.

> Prototype / concept stage. Not production-ready. Forked from
> [`sysnet4admin/notiflex-platform`](https://github.com/sysnet4admin/notiflex-platform).

## The four layers, mapped

| Layer | What it removes | Where it lives here |
| --- | --- | --- |
| 1. Plans humans read | missing context | [`docs/`](docs/) including 16 recorded decisions ([`architecture-decisions.md`](docs/architecture-decisions.md)) |
| 2. Context distilled for the AI | stale memory | [`claude-context/`](claude-context/) |
| 3. Command guardrails | improvised steps | [`CLAUDE.md`](CLAUDE.md) |
| 4. Locked values | interpretation | [`helm-values/`](helm-values/), [`k8s/`](k8s/), [`argocd/`](argocd/) |

## Try it: ask the repository

Clone this repo, run Claude Code inside it, and ask in natural language.
No cluster required for these two:

```text
1. Why was Alertmanager chosen over Grafana Alerting?
   -> answered from the recorded decisions, alternatives included

2. Create an onboarding guide for a new engineer.
   -> generated from the repo; compare with docs/onboarding.md
```

## Links

- Organization: [github.com/gitaiops](https://github.com/gitaiops)
- Book repository (hands-on): [`gitaiops/_Book_GitAIOps`](https://github.com/gitaiops/_Book_GitAIOps)
- Author: [Hoon Jo](https://github.com/sysnet4admin) — CNCF & AAIF Ambassador | KubernetesLab | AI & Cloud Native Engineer

---

<div align="center">
<sub>Seen at a conference? This is the repository from the GitAIOps talk demo.</sub>
</div>
