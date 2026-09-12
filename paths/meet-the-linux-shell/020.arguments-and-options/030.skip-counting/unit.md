---
title: Odd numbers only
vars:
  ODD: { pick: ["9", "11", "13", "15"] }
tasks:
  counted:
    check: |
      wait_exec "(^|/)seq 1 2 ${ODD}\$"
    hint: |
      echo "Pass three arguments in this order: the start (1), the step (2), and the stop (${ODD})."
    solve: |
      seq 1 2 $ODD
---

With **three** arguments, `seq` reads them as *start*, *step*, and
*stop*. So a step of 2 skips every other number.

Print the odd numbers from **1** to **${ODD}**.

::task
#active
Waiting for 1, 3, 5, ... up to ${ODD}...
#completed
Those are the odd numbers up to ${ODD}.
::

::tip{title="Every command has its own rules"}
The shell only splits the line into words and hands them over. What
the words mean is up to the command. With `seq`, the same position
means a different thing depending on how many arguments you give.
Other commands have their own rules, and the manual of each command
is where you look them up.
::
