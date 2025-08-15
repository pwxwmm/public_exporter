#!/usr/bin/env python
# -*- coding: utf-8 -*-
# auth: mmwei3
# version: 1.2.0
# date: 2025-03-07
# python2

import subprocess
import socket
import time
from datetime import datetime, timedelta

class ExecCmdTimeout:
    def __init__(self, cmd, timeout=20):
        self.cmd = cmd
        self.sub = subprocess.Popen(
            self.cmd,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            shell=True,
            bufsize=4096,
        )
        end_time = datetime.now() + timedelta(seconds=timeout)
        while True:
            if self.sub.poll() is not None:
                break
            time.sleep(1)
            if datetime.now() >= end_time:
                self.sub.kill()

    def getoutput(self):
        try:
            output = self.sub.stdout.read().split("\n")
            return [line.strip() for line in output if line.strip()]
        except Exception:
            return []

def parse_time(time_str):
    try:
        time_str = time_str.split("LINK")[0].strip()
        return datetime.strptime(time_str, "%a %b %d %H:%M:%S %Y")
    except Exception:
        return None

def get_system_boot_time():
    try:
        with open("/proc/uptime", "r") as f:
            uptime_seconds = float(f.readline().split()[0])
        boot_time = datetime.now() - timedelta(seconds=uptime_seconds)
        return boot_time
    except Exception:
        return None

def total_seconds(delta):
    return delta.days * 86400 + delta.seconds + delta.microseconds / 1e6

def check_optical_link_status():
    metrics = []
    boot_time = get_system_boot_time()
    boot_threshold = boot_time + timedelta(minutes=15) if boot_time else None

    for i in range(8):
        cmd = "hccn_tool -i {} -link_stat -g".format(i)
        res = ExecCmdTimeout(cmd, timeout=60).getoutput()

        if len(res) < 3:
            metrics.append('optical_link_count{{id="{}"}} 0.1'.format(i))
            metrics.append('optical_link_time{{id="{}"}} 0.1'.format(i))
            continue

        try:
            link_down_count = 0
            last_down_time = None
            last_up_time = None

            for line in res:
                if "link down count" in line.lower():
                    link_down_count = int(line.split(":")[-1].strip())
                elif "LINK DOWN" in line:
                    last_down_time = parse_time(line.split("]")[1])
                elif "LINK UP" in line:
                    last_up_time = parse_time(line.split("]")[1])
                if last_down_time and last_up_time:
                    break

            if last_down_time and boot_threshold and last_down_time < boot_threshold:
                last_down_time = None
            if last_up_time and boot_threshold and last_up_time < boot_threshold:
                last_up_time = None

            optical_link_count = link_down_count
            optical_link_time = 0
            if last_down_time and last_up_time:
                optical_link_time = abs(total_seconds(last_up_time - last_down_time))

            metrics.append('optical_link_count{{id="{}"}} {}'.format(i, optical_link_count))
            metrics.append('optical_link_time{{id="{}"}} {}'.format(i, optical_link_time))

        except Exception:
            metrics.append('optical_link_count{{id="{}"}} 0.1'.format(i))
            metrics.append('optical_link_time{{id="{}"}} 0.1'.format(i))
            continue

    data = "\n".join(metrics)
    print data

def main():
    check_optical_link_status()

if __name__ == "__main__":
    main()
