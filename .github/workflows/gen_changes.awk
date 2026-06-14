NR==FNR {
    old[$1] = $2
    next
}
{
    new[$1] = $2
    if ($1 in old) {
        if (old[$1] != $2) {
            updates[$1] = old[$1] " | " $2
        }
    } else {
        additions[$1] = $2
    }
}
END {
    for (pkg in old) {
        if (!(pkg in new)) {
            removals[pkg] = old[pkg]
        }
    }

    if (image != "") {
      print "#### " image
    }

    if (!length(updates)   &&
        !length(additions) &&
        !length(removals)) {
        print "##### *No package changes*"
        exit
    }

    printf "##### *Package changes (%d updated, %d added, %d removed)*\n", length(updates), length(additions), length(removals)
    print "<details>"
    print ""
    print "| Change | Package | Old Version | New Version |"
    print "| ------ | ------- | ----------- | ----------- |"

    if (length(updates) > 0) {
        for (pkg in updates) {
            printf "| Updated | %s | %s |\n", pkg, updates[pkg]
        }
    }

    if (length(additions) > 0) {
        for (pkg in additions) {
            printf "| Added | %s | - | %s |\n", pkg, additions[pkg]
        }
    }

    if (length(removals) > 0) {
        for (pkg in removals) {
            printf "| Removed | %s | %s | - |\n", pkg, removals[pkg]
        }
    }

    print "</details>\n"
}
