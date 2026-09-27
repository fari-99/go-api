# 💬 WhatsApp Module

Pairs a WhatsApp account to the server using [`whatsmeow`](https://github.com/tulir/whatsmeow) (the same multi-device protocol WhatsApp Web uses). The linked session is stored in Postgres (`sqlstore`) so it survives restarts — once paired, the client auto-reconnects on boot without needing a new QR code.

There are two ways to pair a device: watch the QR directly in the server terminal, or drive it entirely over the REST API (what the Antigravity dashboard's WhatsApp page uses).

---

## 🖥️ Option 1 — Terminal

Set the following in `.env` before starting the service:

```bash
SHOW_QR_CODE_TERMINAL=true
```

Then start the app and trigger a login. The QR will be rendered directly in the terminal as ASCII blocks (via `qrterminal`) in addition to being cached in Redis for the REST endpoints below.

```bash
go run cmd/servers/main/main.go
```

Trigger pairing (requires an authenticated request — see [Auth](#-auth)):

```bash
curl -X POST http://go-api.fadhlan.loc/whatsapp/login \
  -H "Authorization: Bearer <access-token>"
```

Watch the server logs — a QR block will print within a couple of seconds. Scan it with **WhatsApp → Settings → Linked Devices → Link a Device**. It refreshes automatically roughly every 20 seconds until scanned or the 60 second window closes.

---

## 🌐 Option 2 — REST API

All routes live under `/whatsapp` and require the same JWT auth as the rest of the API.

### `POST /whatsapp/login`
Starts pairing. No-ops if a device is already paired (`IsLoggedIn()` true).

```bash
curl -X POST http://go-api.fadhlan.loc/whatsapp/login \
  -H "Authorization: Bearer <access-token>"
```
```json
{
  "status": 200,
  "success": true,
  "data": { "message": "QR generation triggered, poll GET /whatsapp/qr-code to retrieve it" }
}
```

### `GET /whatsapp/qr-code`
Returns the current QR code as a PNG image (`Content-Type: image/png`) — ready to drop straight into an `<img>` tag. 404s if no code is cached (not initiated, already paired, or expired).

```bash
curl http://go-api.fadhlan.loc/whatsapp/qr-code \
  -H "Authorization: Bearer <access-token>" \
  -o qr.png
```

The code is refreshed in Redis every time whatsmeow rotates it (~20s), so keep polling this endpoint (every 2–3s is reasonable) while waiting for a scan — don't cache the image client-side for the full 60s TTL.

### `GET /whatsapp/status`
Reports whether the client is actually paired and usable — not just whether the websocket is open.

```bash
curl http://go-api.fadhlan.loc/whatsapp/status \
  -H "Authorization: Bearer <access-token>"
```
```json
{
  "status": 200,
  "success": true,
  "data": {
    "connected": true,
    "logged_in": true,
    "jid": "6281234567890:1@s.whatsapp.net"
  }
}
```

### `POST /whatsapp/logout`
Unlinks the device (if paired) and resets the client so a fresh QR can be generated. Also cleans up a stale, unpaired socket left over from an abandoned pairing attempt.

```bash
curl -X POST http://go-api.fadhlan.loc/whatsapp/logout \
  -H "Authorization: Bearer <access-token>"
```

### Typical flow
1. `POST /whatsapp/login`
2. Poll `GET /whatsapp/qr-code` every 2–3s and render the PNG until it succeeds
3. Poll `GET /whatsapp/status` in parallel — once `connected: true`, stop polling and show the paired state
4. `POST /whatsapp/logout` any time to unlink and start over

This is exactly what [`WhatsappManager.vue`](../../../vue-frontend/src/components/WhatsappManager.vue) in the dashboard does.

---

## 🔐 Auth
Every route requires the standard `Authorization: Bearer <access-token>` JWT used across the API — see the top-level [README](../../README.md) for how to obtain one.

## 📝 Notes
- The linked device shows up as **Chrome** in WhatsApp's Linked Devices list (configured in [`configs/whatsapp.go`](../configs/whatsapp.go) via `store.DeviceProps`) instead of the `whatsmeow` library default.
- Pairing state lives in Postgres, independent of Redis — Redis only caches the transient QR code and its TTL.
- `whatsmeow` itself is very chatty at `DEBUG` (websocket frames, IQ stanzas, every QR rotation). Set `WHATSAPP_LOG_LEVEL` in `.env` to `WARN` (default) or `ERROR` to quiet it down, or back to `DEBUG` when actually troubleshooting the client.
