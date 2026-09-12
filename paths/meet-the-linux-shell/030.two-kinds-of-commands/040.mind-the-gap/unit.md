---
title: Mind the gap
tasks:
  gap_kept:
    timeout: 45
    check: |
      wait_exec --argc 2 '(^|/)echo mind {2,}the {2,}gap$'
    hint: |
      echo "Quote the whole phrase so all three words, double spaces included, form ONE argument to /bin/echo."
    solve: |
      /bin/echo 'mind  the  gap'
---

It is time to combine the last two lessons. Try echoing a phrase with
extra breathing room between the words (two spaces each):

```
echo mind  the  gap
```

The gaps collapse! The shell split the phrase into three arguments,
because whitespace is a separator and *how much* of it never matters,
and `echo` printed the arguments back with single spaces between.

You already know the fix from the `date` recipe: quotes keep a spaced
value together as one argument. Whenever an argument must contain
whitespace, such as a phrase or a file name with a space, quotes are
the answer. Now make the standalone `/bin/echo` print `mind  the  gap`
with the double spaces intact.

::task
#active
Waiting for `/bin/echo` to receive the whole gapped phrase as a single
argument...
#completed
The phrase arrived as one argument with the gaps preserved.
::
