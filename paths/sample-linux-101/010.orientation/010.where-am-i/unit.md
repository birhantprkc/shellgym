---
title: Find out where you are
requires: [readline]
tasks:
  located:
    check: |
      wait_line '^pwd( |$)'
    hint: |
      echo "Type pwd and press Enter to print the directory you are standing in."
    solve: |
      pwd
---

Every shell always sits in some directory, called the working directory.
The `pwd` command (print working directory) prints it. Try it in the
terminal:

```
pwd
```

You should see your home directory: the place you start out in and can
always return to. Checking where you stand is the first thing to do
whenever you are unsure where a command will take effect.

::task
#active
Waiting for you to print your working directory with `pwd`...
#completed
That is your home directory, and your first completed task.
::
