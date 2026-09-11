---
title: Count files with a pipe
variant: haystack=files
vars:
  EXT: { pick: [log, bak, tmp] }
init:
  - name: seed_inventory
    run: |
      rm -rf /tmp/inventory
      mkdir -p /tmp/inventory
      COUNT=$((RANDOM % 15 + 6))
      echo "$COUNT" > /tmp/inventory_expected
      for i in $(seq 1 "$COUNT"); do
        touch "/tmp/inventory/report-$i.$EXT"
      done
      for i in $(seq 1 $((RANDOM % 10 + 4))); do
        touch "/tmp/inventory/notes-$i.txt" "/tmp/inventory/config-$i.cfg"
      done
      chmod -R a+rX /tmp/inventory
tasks:
  counted:
    check: |
      EXPECTED=$(cat /tmp/inventory_expected)
      wait_file_contains "$GYM_USER_HOME/count.txt" "^${EXPECTED}\s*$"
    hint: |
      if [ -f "$GYM_USER_HOME/count.txt" ]; then
        echo "count.txt exists but the number inside is not right. List /tmp/inventory, keep only the names ending in .${EXT} (grep can filter, and a dollar sign anchors a pattern to the end of the line), and count what is left with wc -l."
      else
        echo "List the directory, filter the listing with grep, count the matching lines, and redirect the resulting number into ~/count.txt."
      fi
    solve: |
      ls /tmp/inventory | grep "\.$EXT$" | wc -l > ~/count.txt
---

The directory `/tmp/inventory` is a jumble of files. Count how many of them
have the `.${EXT}` extension and save that number into `~/count.txt`.

Counting by eye is error-prone, so let the tools do it: `ls` prints one
name per line, `grep` keeps only the lines that match, `wc -l` counts the
lines it receives, and `|` passes the output of each step to the next.
Finish with a `>` redirect to capture the number.

::task{name="counted"}
#active
Waiting for the correct count in `~/count.txt`...
#completed
Correct. `list | filter | count > file` is the same pattern as with any
other listing: processes, packages, connections.
::

::hint
---
title: Matching the end of a name
---
A `grep` pattern matches anywhere in the line, so `log` would also match
`catalog.txt`. Anchor it: `\.${EXT}$` means "a literal dot, the extension,
then the end of the line".
::
