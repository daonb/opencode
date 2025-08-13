---
name: "git-status"
description: "Show git status with custom formatting"
trigger: ["status", "st"]
---

# Git Status Command

This command shows the current git status with enhanced formatting.

## Current Status

!git status --porcelain

## Summary

!echo "Files changed: $(git status --porcelain | wc -l)"
!echo "Branch: $(git branch --show-current)"

## Additional Arguments

You provided: $ARGUMENTS
