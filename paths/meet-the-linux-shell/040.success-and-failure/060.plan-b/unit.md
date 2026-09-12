---
title: Plan B
requires: [readline]
vars:
  BOGUS: { pick: [carrots, downhill, moonwalk] }
tasks:
  fallback:
    timeout: 45
    check: |
      LINE=$(wait_line --latest "^date +--${BOGUS} *(&&|;|\|\|) *whoami$") || exit 1
      case "$LINE" in
        *'||'*) ;;
        *';'*) hint_exit "A semicolon prints the user name whether date failed or not. The assignment wants the operator that runs the fallback only on failure." ;;
        *) hint_exit "With && the fallback runs only when date succeeds, and this date is doomed. The assignment wants the operator that runs it on failure." ;;
      esac
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
Plan A complained and plan B answered.
::
