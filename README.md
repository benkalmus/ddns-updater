# Automatic Dynamic DNS updater for namecheap.com

## **Features**

- [x] **Automatically fetches external IP & updates Namecheap DNS**
- [x] **Supports multiple hosts dynamically (`@`, `ai`, `api`, etc.)**
- [x] **Runs with cron or systemd for scheduling**
- [x] **Prevents unnecessary network requests by tracking last IP change**

## **Setup env**

Create a `.env` file in the same directory as the script:

```
DOMAIN_NAME=yourdomain.tld
DDNS_PASSWORD=your_ddns_password
HOSTS=@,ai,api
```

This allows us to **easily add more hosts** (`ai`, `api`, etc.) without modifying the script.

---

## **Making It Executable**

1. Install dependencies:
   ```sh
   go mod tidy
   ```
2. Build:

   ```sh
   go build -o ddns-updater ddns-updater.go
   ```

3. Make it executable:
   ```sh
   chmod +x ddns-updater
   ```

---

## **Setting Up a Cron Job**

To schedule the script every 5 minutes:

```sh
crontab -e
```

Add:

```
# every 15 mins
*/15 * * * * /path/to/ddns-updater >> logs/ddns-updater.log 2>&1
```

---

## **Alternative: Using Systemd**

If you prefer `systemd` over cron:

Create a systemd service file `/etc/systemd/system/ddns-updater.service`:

```ini
[Unit]
Description=Dynamic DNS Updater for Namecheap
After=network.target

[Service]
ExecStart=/path/to/ddns-updater
Restart=always
User=root
StandardOutput=append:/var/log/ddns-updater.log
StandardError=append:/var/log/ddns-updater.log

[Install]
WantedBy=multi-user.target
```

Enable and start the service:

```sh
sudo systemctl daemon-reload
sudo systemctl enable ddns-updater
sudo systemctl start ddns-updater
```

---

## TODO

- Add logrotate to logfile
- Use Slog and timestamp
