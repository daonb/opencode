---
name: "build-project"
description: "Build the project with optional target"
trigger: ["build", "b"]
---

# Build Project

Building project with target: $ARGUMENTS

## Pre-build checks

!echo "Checking dependencies..."
!npm list --depth=0 2>/dev/null || echo "Some dependencies may be missing"

## Build process

!echo "Starting build process..."
!npm run build $ARGUMENTS

## Post-build summary

!echo "Build completed!"
!ls -la dist/ 2>/dev/null || echo "No dist directory found"
