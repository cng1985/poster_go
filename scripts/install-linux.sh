#!/bin/sh
set -eu

APP_NAME=poster
SCRIPT_DIR="$(CDPATH= cd -- "$(dirname "$0")" && pwd)"
PREFIX="${PREFIX:-/opt/poster}"
DATA_DIR="${DATA_DIR:-/var/lib/poster}"
SERVICE_USER="${SERVICE_USER:-poster}"

if [ "$(id -u)" -ne 0 ]; then
    echo "需要管理员权限" >&2
    exit 1
fi

arch="$(uname -m)"
case "$arch" in
    x86_64) bin_arch=amd64 ;;
    aarch64|arm64) bin_arch=arm64 ;;
    *) echo "不支持的架构: $arch" >&2; exit 1 ;;
esac

if [ -x "$SCRIPT_DIR/poster" ]; then
    source_bin="$SCRIPT_DIR/poster"
    config_src="$SCRIPT_DIR/configs"
else
    source_bin="$(dirname "$SCRIPT_DIR")/release/linux-$bin_arch/poster"
    config_src="$(dirname "$SCRIPT_DIR")/configs"
fi
if [ ! -f "$source_bin" ]; then
    echo "找不到 linux/$bin_arch 二进制，请先构建发布包" >&2
    exit 1
fi

id "$SERVICE_USER" >/dev/null 2>&1 || useradd --system --home-dir "$DATA_DIR" --shell /usr/sbin/nologin "$SERVICE_USER"
install -d -o "$SERVICE_USER" -g "$SERVICE_USER" -m 0750 "$DATA_DIR"
install -d -m 0755 "$PREFIX/bin" "$PREFIX/configs"

if [ -x "$PREFIX/bin/poster" ]; then
    systemctl stop "$APP_NAME" || true
    cp -a "$PREFIX/bin/poster" "$PREFIX/bin/poster.prev"
fi

install -m 0755 "$source_bin" "$PREFIX/bin/poster"
if [ ! -f "$PREFIX/configs/application.yaml" ]; then
    if [ ! -d "$config_src" ]; then
        echo "找不到配置模板 $config_src" >&2
        exit 1
    fi
    cp -a "$config_src/." "$PREFIX/configs/"
fi

cat > "/etc/systemd/system/$APP_NAME.service" <<EOF
[Unit]
Description=Poster Studio
After=network.target

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_USER
WorkingDirectory=$PREFIX
ExecStart=$PREFIX/bin/poster --profile pro --config-dir $PREFIX/configs --data-dir $DATA_DIR
Restart=on-failure
RestartSec=3
NoNewPrivileges=true

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
if systemctl is-enabled --quiet "$APP_NAME"; then
    systemctl restart "$APP_NAME"
else
    systemctl enable --now "$APP_NAME"
fi
systemctl --no-pager --full status "$APP_NAME"
echo "查看日志: journalctl -u $APP_NAME"
