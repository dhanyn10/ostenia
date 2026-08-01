#!/bin/sh
# get_resource_usage.sh - Extracts raw CPU, memory, and root disk metrics from the remote host.

# Read initial CPU status
cat /proc/stat

# Pause briefly to capture delta CPU metrics
sleep 0.1

# Read second CPU status
cat /proc/stat

# Read system memory details
cat /proc/meminfo

# Read root disk utilization in Megabytes
df -m /
