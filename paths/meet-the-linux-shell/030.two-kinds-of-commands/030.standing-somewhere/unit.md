---
title: Where are you standing?
tasks:
  twin_ran:
    check: |
      wait_exec '^(/usr)?/bin/pwd$'
    hint: |
      echo "Run the builtin first with: pwd. Then run its twin by full location: /bin/pwd."
    solve: |
      pwd
      /bin/pwd
---

Another famous builtin is `pwd`. Every shell is always "standing" in
some place on the machine, called its **working directory**, and `pwd`
prints where. (Moving around is a topic for a later path. Today, just
take a look.)

Run `pwd` to see where you stand. Like `echo`, it has a standalone twin
at `/bin/pwd`. Run the twin too, and check that both agree.

::task
#active
Waiting for the standalone `/bin/pwd` to report your location...
#completed
Both twins agree on where you stand.
::
