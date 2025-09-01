#!/bin/bash

go build -o publish-test cmd/tasks/commands.go && ./publish-test pt --queue-name test