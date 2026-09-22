#!/usr/bin/env python3
"""Fill a development database with plausible performance data.

For a lab, not for anything real: it writes directly into
statusengine_perfdata, which belongs to the worker. The interface never
does that - a migration test fails the build if one so much as mentions
those tables. This script is a person with a keyboard, standing outside
that rule on purpose, because a monitoring lab that has been switched
off for a day has no data to draw and the charts have nothing to say.

    python3 tools/dev/seed-perfdata.py


Seven days at one-minute resolution for the services the documentation
screenshots show. Values are shaped rather than random: a daily rhythm,
some noise, and the occasional spike, because a flat line and white
noise both look like something is broken.

It deletes its own range before writing, so running it twice does not
double the data.
"""
import math, random, subprocess, time

NOW = int(time.time())
DAYS = 7
STEP = 60
START = NOW - DAYS * 24 * 3600
random.seed(20260922)

def daily(t, peak_hour=14, amount=1.0):
    """0..1, peaking mid-afternoon."""
    hour = (t % 86400) / 3600
    return amount * (0.5 + 0.5 * math.cos((hour - peak_hour) / 24 * 2 * math.pi))

rows = []
def add(host, service, label, t, value, unit):
    rows.append((host, service, label, t, round(value, 4), unit))

for t in range(START, NOW + 1, STEP):
    # Round trip: sub-millisecond on loopback, with rare spikes.
    rta = 0.018 + 0.010 * daily(t) + random.gauss(0, 0.004)
    if random.random() < 0.004:
        rta += random.uniform(0.05, 0.9)
    add('localhost', 'PING', 'rta', t, max(0.008, rta), 'ms')
    # Packet loss: almost always zero, occasionally not.
    loss = 0.0
    if random.random() < 0.002:
        loss = random.choice([10, 20, 30])
    add('localhost', 'PING', 'pl', t, loss, '%')

    # Load average: three series that track each other with lag.
    base = 0.25 + 0.9 * daily(t, peak_hour=11)
    if random.random() < 0.01:
        base += random.uniform(0.5, 2.5)
    add('localhost', 'Current Load', 'load1', t, max(0.01, base + random.gauss(0, 0.18)), '')
    add('localhost', 'Current Load', 'load5', t, max(0.01, base * 0.85 + random.gauss(0, 0.09)), '')
    add('localhost', 'Current Load', 'load15', t, max(0.01, base * 0.7 + random.gauss(0, 0.05)), '')

    # Disk usage: a slow climb with a cleanup halfway through.
    progress = (t - START) / (NOW - START)
    used = 41 + 9 * progress + random.gauss(0, 0.2)
    if progress > 0.55:
        used -= 6
    add('localhost', 'Root Partition', 'used', t, used, '%')

print(f"{len(rows)} rows from {time.ctime(START)} to {time.ctime(NOW)}")

# One multi-row INSERT per chunk; anything bigger hits max_allowed_packet.
CHUNK = 2000
sql = ["DELETE FROM statusengine_perfdata WHERE hostname = 'localhost' "
       "AND service_description IN ('PING', 'Current Load', 'Root Partition') "
       f"AND timestamp_unix >= {START};"]
for i in range(0, len(rows), CHUNK):
    values = ",".join(
        "('{}','{}','{}',{},{},{},'{}')".format(h, s, l, t * 1000, t, v, u)
        for h, s, l, t, v, u in rows[i:i + CHUNK])
    sql.append("INSERT INTO statusengine_perfdata "
               "(hostname, service_description, label, timestamp, timestamp_unix, value, unit) "
               f"VALUES {values};")

subprocess.run(
    ['mysql', '-h127.0.0.1', '-ustatusengine-dev', '-pstatusengine-dev', 'statusengine-dev'],
    input="\n".join(sql), text=True, check=True,
    stderr=subprocess.DEVNULL)
print("written")
