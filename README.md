## 极夜社区自动赚积分

编译方法：

```bash
go build -o topfeel-bot
```

环境变量：

```plaintext
SESSION_ID=xxxxxx
ACCESS_TOKEN=xxxxxx
ENABLE_BROWSER_HEADLESS=true
```

直接使用：

```bash
wget https://github.com/jiehui555/topfeel-bot/releases/latest/download/topfeel-bot-linux-amd64 -O /usr/local/bin/topfeel-bot
chmod +x /usr/local/bin/topfeel-bot
```

Alpine 定时任务：

```bash
rc-update add crond
rc-service crond start

vi /etc/crontabs/root
0 * * * * /usr/local/bin/topfeel-bot
```
