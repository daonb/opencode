---
name: "authors"
description: "Show project contributors and author statistics"
trigger: ["authors", "contributors", "devs"]
---

# Project Authors & Contributors

!git log --format='%an' | sort -u | tr '\n' ' • ' | sed 's/ • $/\n/'

