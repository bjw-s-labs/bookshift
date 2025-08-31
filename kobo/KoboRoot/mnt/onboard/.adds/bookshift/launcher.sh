#!/bin/sh

logger -t "BookShift" -p daemon.warning "Launcher: started"

export BOOKSHIFT_CONFIG_FILE=/mnt/onboard/.adds/bookshift/config.yaml

if [ ! -f "$BOOKSHIFT_CONFIG_FILE" ]; then
  echo "BookShift config file not found, creating default config..."
  logger -t "BookShift" -p daemon.warning "Launcher: BookShift config file not found, creating default config..."
  mkdir -p /mnt/onboard/.adds/bookshift
  cp /mnt/onboard/.adds/bookshift/config.yaml.dist $BOOKSHIFT_CONFIG_FILE
  logger -t "BookShift" -p daemon.warning "Launcher: BookShift config file created"
fi

UNINSTALL=/mnt/onboard/.adds/bookshift/UNINSTALL
if [ -f "$UNINSTALL" ]; then
    echo "$UNINSTALL exists, removing BookShift..."
    logger -t "BookShift" -p daemon.warning "Launcher: BookShift UNINSTALL file located, removing BookShift..."
    /mnt/onboard/.adds/bookshift/bookshift kobo uninstall
    logger -t "BookShift" -p daemon.warning "Launcher: BookShift removed"
else
    echo "Running BookShift..."
    logger -t "BookShift" -p daemon.warning "Launcher: BookShift binary execution started"
    mkdir -p /mnt/onboard/BookshiftLibrary
    /mnt/onboard/.adds/bookshift/bookshift run
    logger -t "BookShift" -p daemon.warning "Launcher: BookShift binary execution finished"
fi
logger -t "BookShift" -p daemon.warning "Launcher: finished"
