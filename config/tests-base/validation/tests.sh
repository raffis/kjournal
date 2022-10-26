#!/bin/ash

export timeout=180
export retry=3

echo "# Find logs from log-generator in no stream mode succeeds"
timeout $timeout /bin/ash -c "until kjournal pods -n kjournal-system log-generator --no-stream -o json --field-selector payload.foobar=foobar | grep dummy; do sleep $retry; done"
