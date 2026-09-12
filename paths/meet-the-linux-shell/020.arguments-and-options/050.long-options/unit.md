---
title: The long way to say it
tasks:
  spelled_out:
    check: |
      wait_exec '(^|/)date --utc$'
    hint: |
      echo "Type two dashes and then the word: --utc. It is one argument with no spaces inside it."
    solve: |
      date --utc
---

Most options have two spellings. You used the short one, `-u`. The same
option can be spelled out as a **long option**: two dashes and a word,
`--utc`.

Print the UTC time again, this time using the long spelling.

::task
#active
Waiting for the UTC time via the long option...
#completed
The output is identical to the short spelling.
::

::tip{title="Which spelling to use?"}
Short options are quicker to type. Long options are easier to read
later, which is why seasoned users prefer the long form in saved
scripts.
::
