---
title: Plan B
vars:
  BOGUS: { pick: [carrots, downhill, moonwalk] }
tasks:
  fallback:
    timeout: 45
    check: |
      wait_exec "(^|/)date --${BOGUS}\$"
      wait_exec '(^|/)whoami$'
    hint: |
      echo "Put it all on one line: the doomed date --${BOGUS}, then ||, then whoami."
    solve: |
      date --$BOGUS || whoami
---

The inverse of `&&` is `||` ("or-or"): the second command runs
**only if the first failed**. It is the shell's built-in plan B.

In one line, attempt `date` with the hopeless option `--${BOGUS}`, and
when it fails, fall back to printing your user name.

::task
#active
Waiting for the doomed `date --${BOGUS}` with a `whoami` fallback, one
line...
#completed
Plan A complained and plan B answered, so one line handled both
outcomes. `&&` and `||` are how one-liners make decisions without you
watching.
::
