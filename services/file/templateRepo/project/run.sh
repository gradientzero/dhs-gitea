#!/bin/bash
set -x  # Enable debug mode to print commands

# Pull data from the DVC remote storage (but do not stop on failure)
dvc pull
if [ $? -ne 0 ]; then
    echo "Warning: dvc pull failed, continuing anyway..."
fi

# Run the DVC experiment
dvc exp run
if [ $? -ne 0 ]; then
    echo "Warning: dvc exp run failed, continuing anyway..."
fi

echo "Experiment script finished!"
