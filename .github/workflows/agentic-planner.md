---
on:
  schedule:
    # Every 30 minutes
    - cron: "*/30 * * * *"
  workflow_dispatch:

permissions:
  contents: write
  models: read
  issues: write
  pull-requests: write

timeout_minutes: 15
---

# Agentic Planner

## Components

@include @lib/output-channels/shared-team-issue.md

## Job Description

@include @lib/sample-jobs/planner.md
