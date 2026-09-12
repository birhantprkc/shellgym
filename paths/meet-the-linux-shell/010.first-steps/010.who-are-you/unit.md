---
title: Ask who you are
tasks:
  asked:
    check: |
      wait_exec '(^|/)whoami$'
    hint: |
      echo "Type the single word whoami and press Enter."
    solve: |
      whoami
---

Click into the terminal so it receives your keystrokes. You are logged in
under some user name. The machine knows which one, and the command
`whoami` (read it as "who am I") asks it.

Type it after the prompt, all in one word, and press **Enter**.

::task
#active
Waiting for you to ask the machine who you are...
#completed
That name is your user account.
::

::tip
If you make a typo, **Backspace** erases the last character, and the
**Left** and **Right** arrow keys move the cursor within the line before
you press Enter.
::
