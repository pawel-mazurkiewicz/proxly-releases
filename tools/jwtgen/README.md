# JWT Generator (jwtgen)

Simple helper to mint admin JWTs that are compatible with the Proxly License Server.

## Build

```bash
cd tools/jwtgen
go build
```

## Usage

```bash
./jwtgen --secret "$(op read 'op://proxly/JWT_ADMIN_SECRET')" \
         --sub pawel \
         --role admin \
         --hours 12
```

Options:

- `--secret` (required) – the value of `JWT_ADMIN_SECRET` used by the server.
- `--sub` – subject claim, defaults to `admin`.
- `--role` – role claim, defaults to `admin`.
- `--hours` – validity period (default 24 hours).

All output is printed to stdout. Copy the generated token into your admin UI or `curl` commands.

> Tip: rotate `JWT_ADMIN_SECRET` periodically. When you do, rebuild tokens with the new secret.
