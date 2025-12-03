#!/bin/bash

# ---- CẤU HÌNH ----
LOCAL_FILE="bin/remiaq_run1-linux"          # File binary local
REMOTE_HOST="root@103.163.118.103"          # IP VPS
REMOTE_PATH="/home/naqservice/remiaq"       # Thư mục trên VPS
PM2_NAME="remiaq"                            # Tên process pm2

SSH_KEY="D:/OTHER/backup/key2025.pem"       # Đường dẫn SSH key
PM2_PATH="/root/.nvm/versions/node/v24.7.0/bin/pm2"

# ---- KIỂM TRA TUỲ CHỌN ----
BUILD=false
for arg in "$@"; do
    if [[ "$arg" == "--b" ]]; then
        BUILD=true
    fi
done

# ---- BUILD NẾU CÓ TUỲ CHỌN ----
if [ "$BUILD" = true ]; then
    echo "===> Xóa file local cũ..."
    rm -f "$LOCAL_FILE"
    if [ $? -ne 0 ]; then
        echo "===> Không thể xóa file $LOCAL_FILE. Thoát!"
        exit 1
    fi
    sleep 3
    echo "===> Build binary cho Linux..."
    GOOS=linux GOARCH=amd64 go build -o "$LOCAL_FILE" ./cmd/server/main.go
    if [ $? -ne 0 ]; then
        echo "===> Build thất bại!"
        exit 1
    fi
fi

# ---- TẠO THƯ MỤC TRÊN VPS ----
echo "===> Kiểm tra / tạo thư mục trên VPS..."
ssh -i "$SSH_KEY" "$REMOTE_HOST" "mkdir -p $REMOTE_PATH"
if [ $? -ne 0 ]; then
    echo "===> Tạo thư mục thất bại!"
    
fi
ssh -i "$SSH_KEY" "$REMOTE_HOST" <<EOF
if [ -f "$REMOTE_PATH/$(basename $LOCAL_FILE)" ]; then
    NOW=\$(date +"%d-%m-%Y_%H-%M")
    mv "$REMOTE_PATH/$(basename $LOCAL_FILE)" "$REMOTE_PATH/$(basename $LOCAL_FILE)_\$NOW"
    echo "===> Đổi tên file trên server thành: $(basename $LOCAL_FILE)_\$NOW"
fi
EOF

sleep 1
# ---- UPLOAD FILE ----
echo "===> Upload file bằng SCP..."
scp -i "$SSH_KEY" "$LOCAL_FILE" "$REMOTE_HOST:$REMOTE_PATH/"
if [ $? -ne 0 ]; then
    echo "===> Upload thất bại!"
    exit 1
fi

# ---- SSH VÀ CHMOD + RESTART PM2 ----
echo "===> SSH vào VPS để chmod + restart PM2..."
ssh -i "$SSH_KEY" "$REMOTE_HOST" <<EOF
export NVM_DIR="\$HOME/.nvm"
[ -s "\$NVM_DIR/nvm.sh" ] && \. "\$NVM_DIR/nvm.sh"  # load nvm

# chmod file thực thi
chmod +x $REMOTE_PATH/$(basename $LOCAL_FILE)

# restart pm2
$PM2_PATH restart $PM2_NAME
$PM2_PATH save
EOF

echo "===> Deploy hoàn tất!"
