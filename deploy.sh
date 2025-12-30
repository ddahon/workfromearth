#!/bin/bash

set -e

if [ $# -ne 2 ]; then
    echo "Usage: $0 <user@server> <target>"
    echo "  target: server, scraper, or migrate"
    exit 1
fi

SSH_CONNECTION="$1"
TARGET="$2"

if [ "$TARGET" != "server" ] && [ "$TARGET" != "scraper" ] && [ "$TARGET" != "migrate" ]; then
    echo "Error: target must be 'server', 'scraper', or 'migrate'"
    exit 1
fi

if [ "$TARGET" = "migrate" ]; then
    echo "Deploying migrations to $SSH_CONNECTION..."
    
    # Ensure data directory exists with proper permissions
    ssh "$SSH_CONNECTION" "sudo mkdir -p /opt/workfromearth/data && sudo chown wfe:wfe /opt/workfromearth/data && sudo chmod 700 /opt/workfromearth/data"
    
    echo "Copying migration files..."
    scp -r database/migrations "$SSH_CONNECTION:/tmp/migrations_$$"
    ssh "$SSH_CONNECTION" "sudo rm -rf /opt/workfromearth/data/migrations && sudo mv /tmp/migrations_$$ /opt/workfromearth/data/migrations && sudo chown -R wfe:wfe /opt/workfromearth/data/migrations"
    
    MIGRATE_SCRIPT="/opt/workfromearth/migrate.sh"
    TEMP_SCRIPT="/tmp/migrate_$$.sh"
    scp migrate.sh "$SSH_CONNECTION:$TEMP_SCRIPT"
    ssh "$SSH_CONNECTION" "sudo mv $TEMP_SCRIPT $MIGRATE_SCRIPT && sudo chmod 755 $MIGRATE_SCRIPT && sudo chown wfe:wfe $MIGRATE_SCRIPT"
    
    # Ensure database file has proper permissions (if it exists)
    DB_PATH="/opt/workfromearth/data/db.sqlite"
    ssh "$SSH_CONNECTION" "if [ -f $DB_PATH ]; then sudo chown wfe:wfe $DB_PATH && sudo chmod 644 $DB_PATH; fi"
    
    echo "Running migrations on remote server..."
    ssh "$SSH_CONNECTION" "sudo -u wfe $MIGRATE_SCRIPT $DB_PATH"
    
    echo "Migration deployment complete!"
else
    echo "Building $TARGET..."
    make "$TARGET"
    
    BINARY_PATH="bin/$TARGET"
    
    if [ ! -f "$BINARY_PATH" ]; then
        echo "Error: Binary not found at $BINARY_PATH"
        exit 1
    fi
    
    DEPLOY_PATH="/opt/workfromearth/bin"
    echo "Deploying $BINARY_PATH to $SSH_CONNECTION:$DEPLOY_PATH/"
    
    ssh "$SSH_CONNECTION" "sudo mkdir -p $DEPLOY_PATH && sudo chown wfe:wfe $DEPLOY_PATH"
    
    TEMP_PATH="/tmp/${TARGET}_$$"
    scp "$BINARY_PATH" "$SSH_CONNECTION:$TEMP_PATH"
    
    ssh "$SSH_CONNECTION" "sudo mv $TEMP_PATH $DEPLOY_PATH/$TARGET && sudo chmod 755 $DEPLOY_PATH/$TARGET && sudo chown wfe:wfe $DEPLOY_PATH/$TARGET"
    
    echo "Deployment complete! Binary deployed to $DEPLOY_PATH/$TARGET"
fi

