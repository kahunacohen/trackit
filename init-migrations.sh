#!/usr/bin/env bash
# A script to initialize the schema using atlas. Only meant to
# be run once, but it's in a script for purposes of history.
atlas migrate diff initial --dir file://internal/db/migrations --to file://internal/db/schema.sql --dev-url "sqlite://dev?mode=memory" --format '{{ sql . " " }}'
