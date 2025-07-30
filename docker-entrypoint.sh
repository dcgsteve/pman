#!/bin/sh

# Adjust user and group IDs if provided via environment variables
if [ "$PMAN_UID" != "987" ] || [ "$PMAN_GID" != "987" ]; then
    echo "Adjusting pman user: UID=$PMAN_UID, GID=$PMAN_GID"
    
    # Change group ID
    groupmod -g "$PMAN_GID" pman
    
    # Change user ID and group
    usermod -u "$PMAN_UID" -g "$PMAN_GID" pman
    
    # Fix ownership of data directory
    chown -R pman:pman /data
fi

# Ensure proper ownership of data directory
chown -R pman:pman /data

# Switch to pman user and execute the command
exec su-exec pman "$@"