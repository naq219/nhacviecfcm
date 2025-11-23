#!/bin/bash

# ---- CẤU HÌNH ----

LOCAL_FILE="bin/remiaq_run1-linux"               # File binary trên máy local
REMOTE_HOST="root@103.163.118.103"                   # Thay IP VPS của bạn
REMOTE_PATH="/home/naqservice/remiaq"  # Đích trên VPS
PM2_NAME="remiaq"                                # Tên process pm2

SSH_KEY="D:/OTHER/backup/key2025.pem"                       # Đường dẫn key SSH của bạn
PM2_PATH="/root/.nvm/versions/node/v24.7.0/bin/pm2"        # đường dẫn đầy đủ pm2

# ---- TRIỂN KHAI ----

echo "===> Upload file bằng SSH key..."
scp -i "$SSH_KEY" "$LOCAL_FILE" "$REMOTE_HOST:$REMOTE_PATH"

echo "===> SSH vào VPS để chmod + restart pm2..."
ssh -i "$SSH_KEY" "$REMOTE_HOST" <<EOF
export NVM_DIR="$HOME/.nvm"
    [ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"    # load nvm
    chmod +x $REMOTE_PATH
    /root/.nvm/versions/node/v24.7.0/bin/node $PM2_PATH restart $PM2_NAME
    /root/.nvm/versions/node/v24.7.0/bin/node $PM2_PATH save
EOF

echo "===> Done!"
