# Proxly License Admin Web Tool

A lightweight single-page web interface for remotely managing the Proxly License Server.

## Features

- Configure the API base URL and admin JWT (persisted locally in the browser).
- List licenses with activation counts and status.
- Create, edit, or delete licenses (max activations, status, metadata JSON).
- View detailed metadata and activation history for a license.
- Reset activation count, removing all historical activation records.

> **Security note:** This tool runs entirely client-side. Only use it from a trusted machine and network. The JWT token remains in your browser's local storage.

## Getting Started

1. Open `index.html` in a modern browser (double-click or serve over `file://`).
2. Click the gear icon to provide:
   - `API Base URL` – e.g. `http://192.168.0.10:3939/v1`
   - `Admin JWT Token` – mint one with `tools/jwtgen` using the server's `JWT_ADMIN_SECRET`.
3. Save the configuration and begin managing licenses.

The UI accepts the same REST contract documented in `proxly-license-server/api/openapi.yaml`, so any reverse proxy or TLS termination used in production can be targeted by adjusting the base URL.

## Development Notes

The app is framework-free vanilla HTML/JS/CSS bundled into a single file, so no build step is required. If you prefer running through a local server (recommended for CORS testing), you can use:

```bash
npx serve tools/license-admin-web
```

This tool requires the server's CORS middleware to allow the origin (the default middleware is permissive). EOF
