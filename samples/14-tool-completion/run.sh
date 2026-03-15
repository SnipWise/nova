#!/bin/bash
docker model pull huggingface.co/menlo/lucy-gguf:Q4_K_M
go run main.go