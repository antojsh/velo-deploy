---
title: Seguridad
description: Cómo Velo aísla las apps y en qué podés confiar.
---

import { Aside } from '@astrojs/starlight/components';

Velo Deploy hereda su modelo de seguridad de systemd. Cada app corre en una unidad endurecida sin acceso al resto del sistema.

## Usuario Linux por app

Cuando se despliega una app, Velo crea un usuario Linux dedicado sin password, sin login shell y sin directorio home:

```
velo-<appname>:x:1001:1001::/nonexistent:/usr/sbin/nologin
```

La app corre bajo este usuario. No puede leer archivos de otros usuarios, y el SO nunca le va a presentar un prompt de login.

## Unidad systemd endurecida

Cada archivo de unidad incluye las siguientes directivas:

| Directiva | Efecto |
| --- | --- |
| `ProtectSystem=full` | `/usr`, `/boot`, `/etc` se montan read-only. |
| `ProtectHome=true` | La app no puede leer `/home`, `/root` ni `/run/user`. |
| `PrivateTmp=true` | La app tiene su propio namespace de `/tmp`. |
| `NoNewPrivileges=true` | La app no puede escalar privilegios con binarios setuid. |
| `ReadWritePaths=` | Los paths writable se limitan al dir de la app y `/tmp`. |
| `PrivateDevices=true` | La app no puede hablar con devices crudos. |
| `ProtectKernelTunables=true` | Se bloquean escrituras a `/proc` y `/sys`. |
| `RestrictNamespaces=true` | La app no puede crear nuevos namespaces. |
| `MemoryDenyWriteExecute=true` | Política de memoria W^X (donde el kernel lo soporte). |

Un ejemplo completo:

```ini
[Unit]
Description=Velo Deploy app: my-api
After=network-online.target

[Service]
Type=simple
User=velo-my-api
Group=velo-my-api
WorkingDirectory=/opt/deploy/apps/my-api
ExecStart=/opt/nvm/versions/node/v20.10.0/bin/node index.js
Restart=on-failure
RestartSec=5

NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=full
ProtectHome=true
ReadWritePaths=/opt/deploy/apps/my-api /tmp
PrivateDevices=true
ProtectKernelTunables=true
RestrictNamespaces=true
MemoryDenyWriteExecute=true

[Install]
WantedBy=multi-user.target
```

## Aislamiento de filesystem

La app puede:

- Leer y escribir `/opt/deploy/apps/<app>/`.
- Leer y escribir `/tmp` (namespace privado).
- Leer todo en `/usr`, `/etc`, `/var` (montados read-only).
- Hacer conexiones de red salientes.

La app **no puede**:

- Leer los homes de otros usuarios.
- Escribir fuera de su propio directorio (excepto su propio `/tmp`).
- Spawnear procesos setuid.
- Montar filesystems.
- Hablar con block devices crudos.

## Aislamiento de red

Caddy es el único proceso escuchando en un puerto público. Las apps Node se bindean a `localhost` únicamente — no son accesibles directo desde la red. Todo el tráfico fluye por el reverse proxy de Caddy, que termina TLS y forwardea al upstream correcto.

## Secretos

<Aside type="warning" title="No commitees secretos a Git">
	Velo Deploy no gestiona secretos. Pasalos como variables de entorno en la unidad systemd, nunca como parte del repositorio.
</Aside>

Para inyectar variables de entorno, editá la unidad y agregá una línea `Environment=`, o usá la directiva `EnvironmentFile=`:

```ini
[Service]
Environment="DATABASE_URL=postgres://..."
EnvironmentFile=/etc/velo-deploy/my-api.env
```

Después recargá y reiniciá:

```bash
sudo systemctl daemon-reload
sudo systemctl restart velo-my-api
```

## Seguridad del webhook

Mirá [Webhooks](/velo-deploy/es/guide/webhooks/#seguridad) para el patrón recomendado de validación HMAC y allow-listing de IPs.

## Supply chain

Las dependencias Go de Velo están pineadas en `go.sum` y se verifican en cada build. El pipeline de release firma los binarios con [cosign](https://github.com/sigstore/cosign) y el script de install chequea la firma antes de instalar. Mirá la [configuración de release-please](https://github.com/antojsh/velo-deploy/blob/master/release-please-config.json) para el pipeline exacto.

## Reportar una vulnerabilidad

Mirá [SECURITY.md](https://github.com/antojsh/velo-deploy/blob/master/SECURITY.md) en GitHub.
