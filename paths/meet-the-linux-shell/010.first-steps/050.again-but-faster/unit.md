---
title: Again, but faster
tasks:
  reran_user:
    check: |
      wait_exec '(^|/)whoami$'
    solve: |
      whoami
  reran_host:
    check: |
      wait_exec '(^|/)hostname$'
    solve: |
      hostname
  reran_tty:
    check: |
      wait_exec '(^|/)tty$'
    solve: |
      tty
---

This is a quick second lap. Run all three commands you have met so far:
the one that prints your user name, the one that prints the machine's
name, and the one that prints this terminal's name. This time, **do not
retype them**.

Press the **Up** arrow until an earlier command reappears, adjust it if
needed, and press Enter.

::task{name="reran_user"}
#active
Waiting for the user-name command again...
#completed
That is one of the three done.
::

::task{name="reran_host"}
#active
Waiting for the machine-name command again...
#completed
That is two of the three done.
::

::task{name="reran_tty"}
#active
Waiting for the terminal-name command again...
#completed
All three are done, hopefully with far fewer keystrokes this time.
::

::tip
Finding a command with **Up** can take many
presses once the history becomes long. Press **Ctrl-R** instead and type a few letters of the command
you want. The shell shows the latest match. Press **Enter** to run it,
or **Ctrl-R** again for an older match.
::
