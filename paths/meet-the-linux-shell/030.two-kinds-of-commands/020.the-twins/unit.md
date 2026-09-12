---
title: The echo twins
tasks:
  twin_ran:
    check: |
      wait_exec '^(/usr)?/bin/echo .+'
    hint: |
      echo "Type the full location /bin/echo, a space, and then any words you like."
    solve: |
      /bin/echo hello from the standalone twin
---

The command `echo` simply prints back whatever arguments you give it:

```
echo have we met before
```

Try it. Then note that `echo` is a **builtin**. When you type it, the
shell launches no program and prints the words itself, instantly. A
standalone program file with the same name *also* exists, at
`/bin/echo`. Typing a command's full location runs that exact file,
bypassing the builtin.

Print a greeting using the standalone twin at `/bin/echo`.

::task
#active
Waiting for a message printed by the standalone `/bin/echo`...
#completed
This time a real separate program printed the greeting.
::

::tip
**Ctrl-A** jumps to the beginning of the line, and **Ctrl-E** jumps to
the end.
::
