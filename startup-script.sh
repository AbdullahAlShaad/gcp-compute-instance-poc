#!/bin/bash

echo "Hello from script" >> hello.txt

METADATA_VALUE=$(curl http://metadata.google.internal/computeMetadata/v1/instance/attributes/envr -H "Metadata-Flavor: Google")

echo "$METADATA_VALUE" >> hello2.txt