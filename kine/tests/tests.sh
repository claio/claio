#!/bin/bash

set -e

cd tests
for file in $(find . -name '*.sh' | grep -v tests.sh); do
    echo "-----------------------------------------"
    echo "Running: $file"
    $file
done