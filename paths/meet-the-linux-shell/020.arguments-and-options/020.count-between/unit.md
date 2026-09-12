---
title: Count between two numbers
vars:
  LO: { pick: ["2", "3", "4"] }
  HI: { pick: ["11", "12", "13"] }
tasks:
  counted:
    check: |
      wait_exec "(^|/)seq ${LO} ${HI}\$"
    hint: |
      echo "This time pass two arguments, first ${LO} and then ${HI}, separated by whitespace."
    solve: |
      seq $LO $HI
---

The `seq` command also accepts **two** arguments: where to start and where to stop.

Count from **${LO}** to **${HI}**. Whitespace separates one argument
from the next, and the shell splits the line on it before `seq` ever
sees the arguments. One space is enough, and extra spaces do no harm.

::task
#active
Waiting for a count from ${LO} to ${HI}...
#completed
That is a count from ${LO} to ${HI}.
::

::tip
There is no need to retype the command. Press the **Up** arrow, edit
the numbers, and press Enter.
::
