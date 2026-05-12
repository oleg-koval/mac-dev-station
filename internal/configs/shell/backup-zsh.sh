#!/bin/bash
# Backup zsh configuration daily
# Called by cron at 9am

set -e

BACKUP_REPO="$HOME/code/zshrc-backups"
BACKUP_DATE=$(date +%Y%m%d-%H%M%S)

# Ensure repo exists
if [ ! -d "$BACKUP_REPO/.git" ]; then
    git clone git@github.com:oleg-koval/zshrc-backups.git "$BACKUP_REPO" || exit 0
fi

# Copy current config
cp ~/.zshrc "$BACKUP_REPO/zshrc-$BACKUP_DATE"

# Stage and commit
cd "$BACKUP_REPO"
git add -A
git commit -m "Backup $BACKUP_DATE" || true
git push origin main || true

# Keep only last 30 backups
ls -t zshrc-* | tail -n +31 | xargs -r rm

exit 0
