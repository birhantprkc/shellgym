---
title: Three commands on one line
tasks:
  rollcall:
    timeout: 45
    check: |
      wait_exec '(^|/)whoami$'
      wait_exec '(^|/)hostname$'
      wait_exec '(^|/)tty$'
    hint: |
      echo "You need one line with three commands (whoami, hostname, tty) and two semicolons between them."
    solve: |
      whoami; hostname; tty
---

You do not have to run commands one prompt at a time. A semicolon `;`
separates several commands on a single line. The shell runs them left
to right, each one after the previous finishes, no matter whether it
succeeded or failed.

Do the full identity roll call (user name, machine name, terminal
name) in **one line**.

::task
#active
Waiting for the one-line roll call: who you are, what machine you are on, on
which terminal...
#completed
One Enter press produced three answers. The `;` is the "and then" of
the shell: it sequences commands blindly and asks no questions. The
next units introduce chaining that *does* ask questions.
::
