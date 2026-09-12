---
title: Mind the gap
requires: [readline]
tasks:
  gap_kept:
    timeout: 45
    check: |
      LINE=$(wait_line --latest '(^|/)echo .*mind.*the.*gap') || exit 1
      # readline reports the line exactly as typed, double spaces and all -
      # the shell only collapses an unquoted run of spaces later, when it
      # splits the line into arguments. So the tell is the quotes, not the
      # spaces: the correct answer wraps the phrase, the naive one does not.
      case "$LINE" in
        *'"mind  the  gap"'*) ;;
        *"'mind  the  gap'"*) ;;
        *) hint_exit "The three words are there, but they are not quoted. Without quotes the shell collapses each double gap before echo ever sees it. Wrap the whole phrase in one pair of quotes." ;;
      esac
    hint: |
      echo 'Run echo followed by the phrase "mind  the  gap", wrapped in quotes so the double spaces are kept together.'
    solve: |
      echo "mind  the  gap"
---

Time to combine what you know about arguments and quoting. Try echoing a
phrase with extra breathing room between the words (two spaces each):

```
echo mind  the  gap
```

The gaps collapse! The shell split the phrase into three separate
arguments, because whitespace is a separator and *how much* of it never
matters, and `echo` printed the arguments back with a single space
between each.

You already know the fix from the `date` recipe: quotes keep a spaced
value together as one argument. Whenever an argument must contain
whitespace, such as a phrase or a file name with a space, quotes are the
answer. Now print `mind  the  gap` again with `echo`, but this time keep
both double gaps intact.

::task
#active
Waiting for `echo` to print the phrase with its double gaps intact...
#completed
The quotes held the phrase together, gaps and all.
::

::tip
Single quotes (`'...'`) and double quotes (`"..."`) both keep spaces
together. They differ in how they treat `$` and a few other characters,
which is a topic for a later path.
::
