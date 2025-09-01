#!/bin/bash

go build -o commands cmd/tasks/commands.go && ./commands qtc --queue-name test