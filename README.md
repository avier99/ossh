# ossh — SSH config and key management, the secure way, by default.

ossh is a terminal UI for managing SSH hosts and keys. Single binary. No account. No cloud. No passwords stored — ever.

---

## Why

Most people use SSH passwords because setting up key auth is a multi-step process with room to get it wrong. ossh makes the secure path the easy path — generating keys with the right settings, copying them to servers correctly, and keeping your `~/.ssh/config` clean and backed up.

---

## What it does

- **Connect** — fuzzy-search your hosts and launch a session
- **Add host** — key auth by default, password auth available with a clear warning
- **Generate key** — ed25519, passphrase prompted, correct permissions, named cleanly
- **Copy key to server** — `ssh-copy-id` with a manual fallback
- **First-run setup** — creates `~/.ssh/` if missing, audits permissions and reports issues

---

## What it won't do

- Store passwords
- Sync to a cloud or require an account
- Touch anything outside `~/.ssh/`
- Replace `ssh` — it wraps native SSH end-to-end. Delete ossh tomorrow and everything it created still works with plain `ssh`.

---

## A note on the name

A few small projects share the name `ossh`. ossh by Synoxis is the security-first one.

