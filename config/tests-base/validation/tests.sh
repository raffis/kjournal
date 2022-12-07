#!/bin/ash

export timeout=300
export retry=3

echo "# Find logs from log-generator in no stream mode succeeds"
timeout $timeout /bin/ash -c "until echo '{\"foobar\": \"foobar\", \"message\": \"dummy log message\"}'; kjournal pods -n kjournal-system validation --watch=false -o json --field-selector payload.foobar=foobar | grep dummy; do sleep $retry; done"
