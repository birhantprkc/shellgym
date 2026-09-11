---
title: Break something (safely)
vars:
  BOGUS: { pick: [pumpkin, sideways, banana, upstairs] }
tasks:
  broke_it:
    check: |
      wait_exec "(^|/)date --${BOGUS}\$"
    hint: |
      echo "Give date the long option --${BOGUS}. It is guaranteed not to exist."
    solve: |
      date --$BOGUS
---

Sooner or later every command you type will fail. Let's get the first
failure over with, on purpose: run `date` with the made-up option
`--${BOGUS}`.

Read the complaint carefully. Error messages follow a pattern worth
recognizing: *who* is complaining (`date:`), *what* it did not like, and
usually a pointer toward help.

::task
#active
Waiting for you to feed `date` the nonsense option `--${BOGUS}`...
#completed
The command refused, explained itself, and the prompt came right back.
Nothing broke: an error message is just the command telling you what
it could not do. (Mistyping a command's *name* is equally harmless, and
the shell itself answers `command not found`.)
::

::tip
If you change your mind halfway through typing a line, **Ctrl-C**
abandons the line and gives you a fresh prompt. Nothing gets run.
::
