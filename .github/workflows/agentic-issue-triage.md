---
on:
  issues:
    types: [opened, reopened]

permissions:
  contents: read
  models: read
  issues: write
  actions: read
  checks: read
  statuses: read
  pull-requests: read

timeout_minutes: 15
---

# Agentic Issue Triage

## Components

@include @lib/output-channels/issue-comment.md

## Job Description

@include @lib/sample-jobs/qa-engineer.md
