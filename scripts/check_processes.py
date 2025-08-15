#!/usr/bin/env python
# -*- coding: utf-8 -*-

from __future__ import print_function
import subprocess
import socket

# Mapping of hostname to list of processes to monitor
HOST_PROCESS_MAP = {
    "jobsub-35-205": ["smbd", "nmbd", "winbind"],
    "jobsub-35-226": ["smbd", "nmbd", "winbind"],
    "jobsub-35-236": ["smbd", "nmbd", "winbind"],
    "jobsub-35-238": ["smbd", "nmbd", "winbind"],
    "sambapdc-64-035": ["smbd", "winbind"],
    "ldapmaster-64-022": ["keepalived"],
    "ldapmaster-64-128": ["keepalived"],
    # Fallback list
    "default": ["public_exporter"],
}


def get_process_list():
    """
    Return process list based on current hostname.
    """
    hostname = socket.gethostname().lower()
    return HOST_PROCESS_MAP.get(hostname, HOST_PROCESS_MAP["default"])


def is_process_running(proc_name):
    """
    Check if the given process is running.
    Returns 1 if running, 0 if not.
    """
    try:
        proc = subprocess.Popen(
            ["pgrep", "-f", proc_name], stdout=subprocess.PIPE, stderr=subprocess.PIPE
        )
        output, _ = proc.communicate()
        return 1 if output.strip() else 0
    except Exception:
        return 0


def main():
    """
    Print the status of each configured process in Prometheus format.
    """
    hostname = socket.gethostname()
    print("# Hostname: {}".format(hostname))

    for proc in get_process_list():
        status = is_process_running(proc)
        print(
            'check_process_status_public_exporter{{name="{}"}} {}'.format(proc, status)
        )


if __name__ == "__main__":
    main()
