#!/bin/bash

# ---- CẤU HÌNH ----
LOCAL_FILE="bin/remiaq_run1-linux"          # File binary local
REMOTE_HOST="root@103.163.118.103"             # IP VPS
REMOTE_PATH="/home/naqservice/remiaq"          # Thư mục trên VPS
PM2_NAME="remiaq"                              # Tên process pm2

SSH_KEY="D:/OTHER/backup/key2025.pem"          # Đường dẫn SSH key
PM2_PATH="/root/.nvm/versions/node/v24.7.0/bin/pm2"

# ---- KIỂM TRA TUỲ CHỌN ----
BUILD=false
for arg in "$@"; do
    if [[ "$arg" == "--build" ]]; then
        BUILD=true
    fi
done

# ---- BUILD NẾU CÓ TUỲ CHỌN ----
if [ "$BUILD" = true ]; then
    echo "===> Build binary cho Linux..."
    GOOS=linux GOARCH=amd64 go build -o "$LOCAL_FILE" ./cmd/server/main.go
    if [ $? -ne 0 ]; then
        echo "===> Build thất bại!"
        exit 1
    fi
fi

# ---- UPLOAD ----
echo "===> Upload file bằng SSH key..."
scp -i "$SSH_KEY" "$LOCAL_FILE" "$REMOTE_HOST:$REMOTE_PATH"

# ---- SSH vào VPS để chmod + restart pm2 ----
echo "===> SSH vào VPS để chmod + restart pm2..."
ssh -i "$SSH_KEY" "$REMOTE_HOST" <<EOF
export NVM_DIR="\$HOME/.nvm"
[ -s "\$NVM_DIR/nvm.sh" ] && \. "\$NVM_DIR/nvm.sh"  # load nvm

# chmod file thực thi
chmod +x $REMOTE_PATH/$(basename $LOCAL_FILE)

# restart pm2
$PM2_PATH restart $PM2_NAME
$PM2_PATH save
EOF

echo "===> Done!"
