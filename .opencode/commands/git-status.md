---
name: "git-status"
description: "Show git status with enhanced formatting"
trigger: ["status", "st", "git-status"]
---

# Git Status Command

This command shows the current git status with enhanced formatting and useful information.

## Current Repository Status

!git status --porcelain

## Branch Information

!echo "Current branch: $(git branch --show-current)"
!echo "Remote tracking: $(git rev-parse --abbrev-ref --symbolic-full-name @{u} 2>/dev/null || echo 'No remote tracking')"

## Summary Statistics

!echo "Files changed: $(git status --porcelain | wc -l | tr -d ' ')"
!echo "Staged files: $(git diff --cached --name-only | wc -l | tr -d ' ')"
!echo "Unstaged files: $(git diff --name-only | wc -l | tr -d ' ')"
!echo "Untracked files: $(git ls-files --others --exclude-standard | wc -l | tr -d ' ')"

## Recent Commits

!echo "Last 3 commits:"
!git log --oneline -3

## Additional Arguments

Arguments provided: $ARGUMENTS
