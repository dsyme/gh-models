---
on:
  schedule:
    # Every 5 minutes
    - cron: "*/10 * * * *"
  workflow_dispatch:

timeout_minutes: 15

permissions:
  contents: write
  models: read
  issues: write
  pull-requests: write
  discussions: write
  actions: read
  checks: read
  statuses: read
---

# Agentic QA Engineer (Freddy)

## Components

@include @lib/output-channels/shared-team-issue.md

## Job Description

@include @lib/sample-jobs/qa-engineer.md
