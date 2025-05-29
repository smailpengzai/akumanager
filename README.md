# akumanager
aku 管理端
请放入到 /opt/akumanager

步骤
1. 编译
>go build -o akumanager main.go
2. 配置
>mkdir /opt/akumanager
将代码放入到 /opt/akumanager 目录下

cat <<EOF > /etc/systemd/system/akumanager.service
[Unit]
Description=aku Manager Service
[Service]
Type=simple
WorkingDirectory=/opt/akumanager
ExecStart=/opt/akumanager/akumanager
Restart=on-failure
[Install]
WantedBy=multi-user.target
EOF

3. 配置开机启动
>systemctl enable akumanager
4. 启动
>systemctl start akumanager
5. 查看日志
>journalctl -u akumanager -f

