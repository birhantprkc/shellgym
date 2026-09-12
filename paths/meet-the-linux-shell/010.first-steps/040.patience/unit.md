---
title: Wait for it
vars:
  NAP: { pick: ["4", "5", "6"] }
tasks:
  napped:
    check: |
      wait_exec "(^|/)sleep ${NAP}s?\$"
    hint: |
      echo "Run sleep with the number ${NAP} as its argument, separated by a space."
    solve: |
      sleep $NAP
---

Some commands take a while to finish, and the prompt only comes back
when they are **done**.

The command `sleep` does nothing on purpose: you give it a number of
seconds, and it simply takes that long. Make this shell sleep for
**${NAP} seconds**. The number is an *argument*, an extra word after
the command name, separated by a space.

::task
#active
Waiting for a ${NAP}-second nap...
#completed
The prompt is back after ${NAP} seconds of silence.
::

::hint{title="Nothing seems to happen?"}
That is the point: while `sleep` runs, the prompt is gone and the shell
is busy. Count to ${NAP} and the prompt will be back.
::

::tip{title="Silence does not mean failure"}
Many commands say nothing when all went well. This is the Unix rule
of silence: a program that has nothing surprising to report prints
nothing, so that its output stays useful as input to other programs.
::
