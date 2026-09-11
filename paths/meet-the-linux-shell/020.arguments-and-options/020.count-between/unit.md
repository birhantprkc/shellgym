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

`seq` also accepts **two** arguments: where to start and where to stop.

Count from **${LO}** to **${HI}**. Whitespace separates one argument
from the next. One space is enough, and extra spaces do no harm.

::task
#active
Waiting for a count from ${LO} to ${HI}...
#completed
The shell split the line into two arguments on whitespace before `seq`
ever saw them. Keep that in mind, because it will matter soon.
::

::tip
There is no need to retype the command. Press the **Up** arrow, edit
the numbers, and press Enter.
::
