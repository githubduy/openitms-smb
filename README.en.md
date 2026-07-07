*English | [Tiếng Việt](README.md)*

# OpenITMS-SMB

**Manage every Windows & Linux machine in your company from a single web page.**

OpenITMS-SMB is open-source software that helps small and medium businesses (SMBs) automate IT work:
install software, apply updates, restart services, and run commands/scripts across many machines at
once — all from a web interface, without logging into each machine.

> **Status: pre-release (in active development).** A packaged installer is on the way. For now you
> can build an installable package from source — see [DEVELOPMENT.md](DEVELOPMENT.md).

---

## What can it do?

- 🖥️ **Remote Windows management** — run PowerShell on Windows hosts over WinRM (certificate auth,
  no passwords), or SSH. Linux machines are supported too.
- ⚡ **WinRS Console** — run a single command on one machine and see the result instantly. It also
  remembers your last run and keeps a history of recent commands.
- 📋 **Task Templates** — define a "job to run" once, then run it again anytime, across many machines.
- ⏰ **Scheduling** — run tasks automatically on a schedule (nightly backups, weekly updates…).
- 🚀 **1-click enrollment** — download a setup script, run it on the target machine; the certificate
  is uploaded automatically and the host appears in your inventory.
- 📦 **Batteries included** — bundles the database (MariaDB), a local git server (Gitea), and
  PowerShell. No Docker, nothing extra to install.
- 🔒 **Secure by default** — passwords/keys are encrypted, with warnings when defaults are still in use.

---

## Installation

You need one server to run OpenITMS — a Windows or Linux machine on your local network. Everyone else
uses it through a browser.

### Linux

```bash
# 1. Download and extract the package
tar -xzf openitms-smb-*-linux-amd64.tar.gz
cd openitms-smb-*-linux-amd64

# 2. Install (single command, needs root)
sudo ./install.sh
```

That's it. The installer sets up everything (database, services, admin account, git server) and
prints the access URL. Requirements: 64-bit Linux with `systemd`. No internet required.

- Uninstall: `sudo ./uninstall.sh`
- Services: `systemctl status openitms openitms-db`

### Windows

Open **PowerShell as Administrator** (right-click → *Run as administrator*):

```powershell
# 1. Extract the package and enter the folder
cd openitms-smb-<version>-windows

# 2. Allow scripts + install (single command)
Set-ExecutionPolicy -Scope Process Bypass -Force
.\installer\windows\install.ps1
```

The installer initializes the database, creates the admin account, sets up the local git server, and
registers OpenITMS to run at startup. It prints the access URL when done. Requirements: 64-bit
Windows 10/11 or Windows Server.

- Uninstall: `.\installer\windows\uninstall.ps1` (add `-PurgeData` to also delete data)
- Status: `Get-Service OpenITMS-DB` and `Get-ScheduledTask OpenITMS*`

---

## Getting started

1. Open a browser to the URL printed after installation (e.g. `http://<server-ip>:3000`).
2. Sign in with the default account **`admin` / `quickwin123`** — **change the password immediately**.
3. A **"Host"** project managing the server itself is already there. Open **WinRS Console** to try it.
4. **Add a Windows machine to manage:** WinRS Console → *Enroll this machine (1-click)* → download the
   script → run it on that machine in PowerShell (Administrator). The host appears in your Inventory.

> 💡 Hover over each menu item to see a short explanation of what it does.

---

## Security notes

- The default `admin / quickwin123` account and the default database password are **known defaults**
  for fast setup — change the admin password right away, and consider changing the DB password
  (environment variable at install time).
- The database listens only locally (localhost); it is not exposed to the internet.
- Certificates live in the server's `certs/` folder; keys/passwords inside the app are encrypted.

---

## Documentation & contributing

- 🛠️ **Build from source / development:** [DEVELOPMENT.md](DEVELOPMENT.md)
- 📚 Full documentation: [docs/](docs/)
- 🤝 Contributing: [CONTRIBUTING.md](CONTRIBUTING.md) · Conduct: [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)

---

## License

MIT — [LICENSE](LICENSE). This is a fork of
[Semaphore UI](https://github.com/semaphoreui/semaphore) © Denis Gukov / Castaway Labs LLC (MIT) —
[LICENSE-SEMAPHORE](LICENSE-SEMAPHORE), [NOTICE.md](NOTICE.md). Not affiliated with or endorsed by
Semaphore UI.
