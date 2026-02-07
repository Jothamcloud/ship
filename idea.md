# Shipfe — Local-First Frontend Deployment CLI (BYOA v1)

## Overview

**Shipfe** is an open-source, local-first CLI tool for deploying frontend applications to AWS with minimal setup.

The goal is simple:

> Run **a few command** from your project root → get a **public URL**  
> Hosted on **AWS (S3 + CloudFront)** using **your own AWS account**  
> No GitHub. No CI/CD. No infrastructure setup.

Shipfe is optimized for **speed, previews, and developer experience**, not complex production pipelines.

---

## Core Philosophy

- **Bring Your Own AWS (BYOA)** only
- Opinionated defaults over flexibility
- Local-first workflow
- No GitHub or CI/CD required
- Same URL on every redeploy
- Built for previews, reviews, and quick testing

Shipfe is **not** a replacement for Terraform, CDK, or production deployment workflows.

---

## Target Users

- Frontend developers who want instant preview URLs
- DevOps / backend engineers deploying simple UIs
- Hackathon and MVP builders
- Teams sharing work for quick review without pushing to GitHub

---

## Non-Goals (Explicitly Out of Scope)

To keep the tool focused, Shipfe **does not support**:

- Server-side rendering (SSR)
- Backend APIs
- Custom CloudFront tuning
- Multi-region deployments
- Custom domains (v1)
- CI/CD integration
- Non-AWS providers

---

## High-Level Workflow

1. User installs Shipfe (single binary)
2. User navigates to a frontend project directory
3. User runs `shipfe deploy`
4. Shipfe:
   - Detects the frontend framework
   - Builds the project
   - Provisions AWS resources (first run only)
   - Uploads build artifacts
   - Returns a CloudFront URL
5. Subsequent deploys reuse the **same infrastructure and URL**

---

## CLI Commands (MVP)

```bash
shipfe deploy      # first deploy or redeploy
shipfe status      # show deployment details
shipfe destroy     # tear down AWS resources
```

---

## Framework Detection (Initial Scope)

Shipfe performs best-effort detection based on project files:

- Vite
- React (CRA)
- Vue
- Static HTML

Detection is opinionated and intentionally limited in v1.

---

## Build Behavior

- Automatically runs the detected build command
- Infers build output directory
- Fails fast on build errors
- No custom build scripting or hooks in v1

---

## AWS Architecture

Shipfe provisions the following resources **in the user’s AWS account**:

### S3

- One private bucket per project
- Stores static build artifacts

### CloudFront

- One distribution per project
- S3 as origin
- Uses Origin Access Control (OAC)

### IAM

- Uses the user’s existing AWS credentials
- Requires only minimal permissions

---

## State Management

Shipfe maintains local project state in:

```text
.shipfe/config.json
```

This file stores:

- AWS region
- S3 bucket name
- CloudFront distribution ID
- Build output path

This enables:

- Stable URLs across redeploys
- Idempotent deployments
- Clean teardown with `shipfe destroy`

---

## Redeploy Behavior

On subsequent `shipfe deploy` runs:

- Rebuild the project
- Sync updated files to S3
- Invalidate CloudFront cache
- Return the **same URL**

No new infrastructure is created.

---

## AWS Credentials

Requirements:

- User must have AWS credentials configured locally
- Shipfe does not manage authentication
- Standard AWS credential resolution is supported (env vars, config files, profiles)

This is a **hard requirement** by design.

---

## Error Handling Principles

- Clear, actionable error messages
- Fail fast on:
  - Missing AWS credentials
  - Unsupported frameworks
  - Build failures

- Never leave partial or orphaned infrastructure silently

---

## Language & Tooling

- Language: **Go**
- AWS SDK: AWS SDK for Go v2
- CLI framework: Cobra (or equivalent)
- Distributed as a single static binary

---

## Future Ideas (Not Implemented in v1)

These are intentionally deferred:

- Ephemeral preview environments
- Hosted (non-BYOA) mode
- Custom domains
- Config overrides
- CI/CD integration

---

## Summary

Shipfe provides the **fastest path** from a local frontend project to a live, AWS-hosted URL using the developer’s own AWS account.

It is:

- Simple
- Opinionated
- Local-first
- AWS-native
- Open source

And designed to stay small, fast, and focused.

```

```
