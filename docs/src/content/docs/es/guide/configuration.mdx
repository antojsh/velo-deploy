---
title: Configuración
description: El schema de config.json, los defaults y los patrones de override.
---

import { Aside } from '@astrojs/starlight/components';

Velo Deploy guarda su estado en un único archivo JSON en `/etc/velo-deploy/config.json`. El archivo es editable a mano; los cambios se toman en la próxima invocación de la CLI.

## Ubicación del archivo

```
/etc/velo-deploy/config.json
```

Override del path con la variable de entorno `VELO_CONFIG`.

## Schema completo

```json
{
  "caddy_conf_dir": "/etc/caddy/conf.d",
  "apps_dir": "/opt/deploy/apps",
  "logs_dir": "/var/log/velo-deploy",
  "daemon_port": "9999",
  "nvm_dir": "/opt/nvm",
  "apps": {
    "my-api": {
      "name": "my-api",
      "type": "node",
      "repo_url": "https://github.com/user/my-api",
      "branch": "main",
      "node_version": "24",
      "port": 3000,
      "domain": "",
      "alias": "my-api.local",
      "node_path": "/opt/nvm/versions/node/v24.0.0/bin/node",
      "entry_point": "index.js"
    },
    "my-site": {
      "name": "my-site",
      "type": "static",
      "domain": "",
      "alias": "my-site.local",
      "output_dir": "dist"
    }
  }
}
```

## Campos de nivel superior

| Campo | Tipo | Default | Descripción |
| --- | --- | --- | --- |
| `caddy_conf_dir` | string | `/etc/caddy/conf.d` | Dónde se escriben los vhosts por app. |
| `apps_dir` | string | `/opt/deploy/apps` | Dónde vive el código fuente de las apps. |
| `logs_dir` | string | `/var/log/velo-deploy` | Dónde se escriben los logs de deploy. |
| `daemon_port` | string | `9999` | Puerto al que se bindea el daemon de webhooks. |
| `nvm_dir` | string | `/opt/nvm` | Ubicación del install de nvm. |
| `apps` | object | `{}` | Mapa de nombre de app a config de app. |

## Campos de app

### Comunes

| Campo | Tipo | Requerido | Descripción |
| --- | --- | --- | --- |
| `name` | string | sí | Identificador único de la app. |
| `type` | `"node"` \| `"static"` | sí | Tipo de runtime. |
| `domain` | string | no | Dominio público. Dispara Let's Encrypt. |
| `alias` | string | no | Alias local para acceso por path. |

### Solo Node.js

| Campo | Tipo | Requerido | Descripción |
| --- | --- | --- | --- |
| `repo_url` | string | sí | URL de Git. |
| `branch` | string | no | Branch a trackear. Default: default del repo. |
| `node_version` | string | no | Major version a usar. Default: `24`. |
| `port` | number | sí | Puerto interno. Default: primer libre en 3000-3999. |
| `node_path` | string | sí | Path absoluto al binario de Node. |
| `entry_point` | string | sí | Archivo a iniciar. |

### Solo estáticos

| Campo | Tipo | Requerido | Descripción |
| --- | --- | --- | --- |
| `repo_url` | string | no | URL de Git. Opcional para `add`. |
| `output_dir` | string | sí | Directorio de output del build. |

## Editar el archivo a mano

<Aside type="caution" title="Validá antes de recargar">
	Después de editar `config.json` a mano, corré `velo-deploy list` para asegurarte de que el JSON parsea. Un archivo roto hace que cada comando falle con error de parseo.
</Aside>

```bash
sudo jq . /etc/velo-deploy/config.json
velo-deploy list
```

Para apps Node.js, reiniciá la unidad afectada después de editar:

```bash
sudo systemctl restart velo-myapp
```

Para sitios estáticos, no se requiere restart — Caddy toma los cambios de vhost en `caddy reload` (que Velo dispara automáticamente).

## Versionado

El formato del config sigue [Semantic Versioning](https://semver.org/). Los campos nuevos son aditivos y vienen con defaults razonables. Renames o removals salen en un release mayor y se anuncian en el [CHANGELOG](https://github.com/antojsh/velo-deploy/blob/master/CHANGELOG.md).
