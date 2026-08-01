#!/bin/sh
# get_cwd.sh - Locates and outputs the active CWD of the latest interactive shell inside the WSL container.

for d in /proc/[0-9]*/; do
    d="${d%/}"
    pid="${d##*/}"

    # Verify the stat file is readable
    if [ -r "/proc/$pid/stat" ]; then
        stat_content=$(cat "/proc/$pid/stat")

        # Extract process name between parentheses
        comm="${stat_content%%)*}"
        comm="${comm##*\(}"

        # Check if process is an interactive shell
        if [ "$comm" = "bash" ] || [ "$comm" = "zsh" ] || [ "$comm" = "sh" ] || [ "$comm" = "fish" ]; then
            fd0=$(readlink "/proc/$pid/fd/0" 2>/dev/null)
            case "$fd0" in
                /dev/pts/*|/dev/tty*)
                    if [ -d "/proc/$pid/cwd" ]; then
                        cwd=$(readlink "/proc/$pid/cwd")
                        if [ -n "$cwd" ]; then
                            echo "$pid $cwd"
                        fi
                    fi
                    ;;
            esac
        fi
    fi
done | sort -n | tail -n 1 | cut -d" " -f2-
