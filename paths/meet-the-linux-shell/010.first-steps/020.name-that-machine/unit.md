---
title: Name that machine
tasks:
  asked:
    check: |
      wait_exec '(^|/)hostname$'
    hint: |
      echo "The command is a single word: hostname. Type it and press Enter."
    solve: |
      hostname
---

Every Linux machine has a name of its own, so that people (and other
machines) can tell them apart. The command that prints it is `hostname`.

Ask this machine for its name.

::task
#active
Waiting for you to ask the machine for its name...
#completed
That is the machine's name.
::

::tip
The **Tab** key completes half-typed command names. Try typing `hostn`
and pressing Tab.
::
