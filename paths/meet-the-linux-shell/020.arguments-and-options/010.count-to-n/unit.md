---
title: Count up
vars:
  LIMIT: { pick: ["7", "9", "12", "14"] }
tasks:
  counted:
    check: |
      wait_exec "(^|/)seq ${LIMIT}\$"
    hint: |
      echo "Type the command name, one space, and the number ${LIMIT}, and nothing else."
    solve: |
      seq $LIMIT
---

The command `seq` counts out loud. Given a single number as its
argument, it counts from 1 up to that number, one per line.

Make it count to **${LIMIT}**.

::task
#active
Waiting for a count from 1 to ${LIMIT}...
#completed
The number was an argument: the shell handed it to `seq`, and `seq`
decided what it meant. That division of labor never changes.
::

::tip
If the screen gets cluttered, **Ctrl-L** wipes it clean. Your command
history stays intact, and only the pixels are cleared.
::
