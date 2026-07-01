---
title: Troubleshooting
description: Diagnosticá los problemas más comunes en producción.
---

import { Aside } from '@astrojs/starlight/components';

## La app no inicia

```bash
sudo systemctl status velo-myapp
sudo journalctl -u velo-myapp -n 200 --no-pager
velo-deploy logs myapp
```

Causas comunes:

- **`node_path` incorrecto** — la `node_version` cambió pero el path no. Editá `/etc/velo-deploy/config.json` y actualizá `node_path`.
- **Falta una dependencia** — `node_modules` quedó vacío después de un `git pull`. Corré `velo-deploy deploy <repo> --force`.
- **Puerto ya en uso** — código de salida `5`. Encontrá el proceso: `sudo ss -tlnp | grep ':3000'`.
- **Permiso denegado** — el dir de la app no es owned por `velo-<app>`. Arreglalo con `sudo chown -R velo-myapp:velo-myapp /opt/deploy/apps/myapp`.

## La app crashea inmediatamente

```bash
sudo journalctl -u velo-myapp -e --no-pager
```

El journal usualmente muestra un stack trace. Causas más comunes:

- Un `await` top-level en un `index.js` sin `type: "module"` configurado.
- Una variable de entorno faltante — mirá [Seguridad → Secretos](/velo-deploy/es/architecture/security/#secretos).
- Un `node_modules` compilado para otra versión de Node después de un upgrade.

## El webhook no dispara

```bash
sudo systemctl status velo-deploy-watcher
sudo journalctl -u velo-deploy-watcher -f
```

Testeá el endpoint localmente:

```bash
curl -X POST http://localhost:9999/webhook
```

Si obtenés connection refused, el daemon no está corriendo. Si obtenés un `404`, la ruta está mal (debería ser `/webhook`, no `/`).

Si GitHub está mandando eventos pero no pasa nada en el servidor, verificá:

- El `repo_url` en `config.json` matchea el URL que GitHub envía.
- La branch es `main` o `master`.
- El user del deploy tiene acceso de escritura a `/opt/deploy/apps/<name>/`.

## HTTPS no funciona

```bash
sudo caddy validate --config /etc/caddy/Caddyfile
sudo caddy reload --config /etc/caddy/Caddyfile
sudo journalctl -u caddy -f
```

Causas comunes:

- **El DNS no apunta al servidor** — `dig +short <domain>` debería devolver la IP del servidor.
- **Puerto 80 bloqueado** — Let's Encrypt necesita el puerto 80 para validación HTTP-01. Verificá con `curl -I http://<domain>/.well-known/acme-challenge/test`.
- **Rate limit** — Let's Encrypt permite 50 certs por semana por dominio. Si lo pegaste, esperá una semana o usá un método de validación distinto.

## Conflictos de puerto

```bash
sudo ss -tlnp | grep -E ':3[0-9]{3}'
```

Si dos apps intentan usar el mismo puerto, el segundo deploy falla con código de salida `5`. Re-desplegá con un `--port` explícito:

```bash
velo-deploy deploy https://github.com/user/api --port 3100
```

## Config vieja de Caddy después de un `remove`

```bash
sudo ls /etc/caddy/conf.d/
sudo rm /etc/caddy/conf.d/<app>.conf
sudo caddy reload --config /etc/caddy/Caddyfile
```

Esto no debería pasar en un `velo-deploy remove` normal, pero si pasa, la limpieza manual es segura.

## Disco lleno

```bash
du -sh /opt/deploy/apps/* | sort -h
du -sh /var/log/velo-deploy/* | sort -h
```

El mayor uso de disco viene de `node_modules`. Para apps Node, podá las dev dependencies en producción:

```bash
velo-deploy deploy https://github.com/user/api
sudo -u velo-user npm prune --production
```

Para sitios estáticos, el output del build suele ser chico.

## Los logs se comen el disco

```bash
sudo journalctl --vacuum-size=100M
```

Esto cape el journal de systemd en 100 MB. Agregá la misma línea a `/etc/systemd/journald.conf`:

```ini
[Journal]
SystemMaxUse=100M
```

Después reiniciá:

```bash
sudo systemctl restart systemd-journald
```

## Pedir más ayuda

<Aside type="tip" title="Antes de abrir un issue">
	Corré `velo-deploy version` y `velo-deploy list --json` e incluí la salida en tu reporte de bug. Eso solo nos dice la mayor parte de lo que necesitamos saber.
</Aside>

- [GitHub Issues](https://github.com/antojsh/velo-deploy/issues) para bugs y feature requests.
- [GitHub Discussions](https://github.com/antojsh/velo-deploy/discussions) para preguntas e ideas.
- [SECURITY.md](https://github.com/antojsh/velo-deploy/blob/master/SECURITY.md) para reportes de vulnerabilidad.
