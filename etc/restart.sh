#!/bin/sh
set -e
cd "$(dirname "$0")"
docker compose pull
docker compose -p fountain-finder up -d --remove-orphans
