---
title: Only if it worked
requires: [readline]
vars:
  PAUSE: { pick: ["2", "3"] }
tasks:
  chained:
    timeout: 45
    check: |
      LINE=$(wait_line --latest "^sleep +${PAUSE}s? *(&&|;|\|\|) *hostname$") || exit 1
      case "$LINE" in
        *'&&'*) ;;
        *';'*) hint_exit "Both commands ran, but a semicolon prints the machine name no matter how the pause ended. The assignment wants the operator that waits for success." ;;
        *) hint_exit "With || the machine name appears only when the pause fails. The assignment wants the operator that waits for success." ;;
      esac
      wait_exec "(^|/)sleep ${PAUSE}s?\$"
      wait_exec '(^|/)hostname$'
    hint: |
      echo "Put it all on one line: the ${PAUSE}-second sleep, then &&, then hostname."
    solve: |
      sleep $PAUSE && hostname
---

Now make the exit status work for you. The operator `&&`
("and-and") chains two commands with a condition: the second one runs
**only if the first succeeded** (exit status `0`). If the first one
fails, the shell skips the second entirely.

In one line, pause for **${PAUSE} seconds**, and then, only if the
pause finished properly, print the machine's name.

::task
#active
Waiting for a ${PAUSE}-second pause followed by the machine name, one
line, chained on success...
#completed
The name appeared only after the pause delivered its `0`.
::
