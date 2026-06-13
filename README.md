# ossh

SSH config and key management, the secure way, by default.

---

## Why

Most people use SSH passwords because setting up key auth is too many steps with too much room to get wrong. ossh makes the secure path the easy path. It handles key generation, copies keys to servers correctly, and keeps your `~/.ssh/config` clean and backed up.

---

## What it does

- **Connect** — fuzzy search your hosts and launch a session
- **Add host** — key auth by default, password auth available with a clear warning
- **Generate key** — ed25519, passphrase prompted, correct permissions, named cleanly
- **Copy key to server** — tries `ssh-copy-id` first, falls back to manual instructions
- **First-run setup** — creates `~/.ssh/` if missing and audits permissions

---

## What it won't do

- Store passwords
- Sync to a cloud or require an account
- Touch anything outside `~/.ssh/`
- Replace `ssh`. It wraps native SSH end-to-end. Delete ossh tomorrow and everything it created still works with plain `ssh`.

---

## A note on the name

A few small projects share the name `ossh`. ossh by Synoxis is the security-first one.

