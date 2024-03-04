#!/bin/bash

# Count successful operations for mixtral
success_mixtral=$(cat results_mixtral.txt | grep ^ok* | wc -l)

# Count failed operations for mixtral
fail_mixtral=$(cat results_mixtral.txt | grep "exit status" | wc -l)

# Count successful operations for gpt4
success_gpt4=$(cat results_gpt4.txt | grep ^ok* | wc -l)

# Count failed operations for gpt4
fail_gpt4=$(cat results_gpt4.txt | grep "exit status" | wc -l)

# Print results
echo "Results Summary:"
echo "----------------"
echo "Mixtral:"
echo "  Successful: $success_mixtral"
echo "  Failed: $fail_mixtral"
echo ""
echo "GPT-4:"
echo "  Successful: $success_gpt4"
echo "  Failed: $fail_gpt4"

