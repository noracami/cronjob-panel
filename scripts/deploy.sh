#!/bin/bash
cd /opt/cronjob-panel
docker compose -f docker-compose.prod.yml pull cronjob-panel 2>&1
docker compose -f docker-compose.prod.yml up -d cronjob-panel 2>&1
echo "Deploy completed at $(date)"
