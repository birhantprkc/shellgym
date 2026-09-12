---
title: Which window is this?
tasks:
  asked:
    check: |
      wait_exec '(^|/)tty$'
    hint: |
      echo "The command is three letters long: tty. It prints something like /dev/pts/0."
    solve: |
      tty
---

You can open several terminal windows to the same machine at once, and
each one is a separate conversation. The command `tty` prints the name of
the terminal *this* conversation is running on.

Run it here. The output looks like `/dev/pts/0`, and a second terminal
window would report a different one.

::task
#active
Waiting for you to ask which terminal this is...
#completed
That is this window's own address.
::
