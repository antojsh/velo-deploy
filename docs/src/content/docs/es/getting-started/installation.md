---
title: Instalación
description: Requisitos del sistema, plataformas soportadas e instalador de una línea.
---

import { Tabs, TabItem, Steps, Aside } from '@astrojs/starlight/components';

Velo Deploy corre en un VPS Linux y orquesta tres servicios del sistema: el binario CLI/TUI `velo-deploy`, el daemon de webhooks `velo-deploy-watcher`, y el servidor web Caddy.

## Requisitos

| Requisito | Mínimo | Recomendado |
| --- | --- | --- |
| SO | Ubuntu 20.04, Debian 11 | Ubuntu 22.04 LTS, Debian 12 |
| Arquitectura | x86_64 (amd64) | x86_64 |
| Privilegios | `root` para la instalación | `root` |
| RAM libre | 512 MB | 1 GB + margen para las apps |
| Disco libre | 5 GB | 20 GB |
| Puertos abiertos | 80, 443 | 80, 443, 9999 (webhooks) |

<Aside type="caution" title="ARM y macOS no están soportados">
	El instalador apunta a `linux/amd64`. El fuente Go compila en otras plataformas, pero la integración con systemd y Caddy solo fue validada en Ubuntu y Debian.
</Aside>

## Instalación de una línea

<Steps>

1. Conectate por SSH a tu VPS como `root`.

	```bash
	ssh root@tu-servidor
	```

2. Ejecutá el instalador. Descarga la última release, valida los checksums y registra la unidad systemd de `velo-deploy-watcher`.

	```bash
	curl -sS https://get.velo-deploy.sh | bash
	```

3. Verificá la instalación.

	```bash
	velo-deploy version
	systemctl status velo-deploy-watcher
	```

</Steps>

## Instalación manual

Usá el flujo manual cuando quieras pinear una versión específica o correr desde el fuente.

<Tabs syncKey="install-method">
	<TabItem label="Desde un tarball" icon="seti:default">

		```bash
		VER=$(curl -sS https://api.github.com/repos/antojsh/velo-deploy/releases/latest | grep tag_name | cut -d '"' -f 4)
		curl -sSLo velo-deploy.tar.gz "https://github.com/antojsh/velo-deploy/releases/download/${VER}/velo-deploy_${VER#v}_linux_amd64.tar.gz"
		tar -xzf velo-deploy.tar.gz
		sudo install -m 0755 velo-deploy /usr/local/bin/velo-deploy
		sudo ./install.sh
		```

	</TabItem>
	<TabItem label="Desde el fuente" icon="seti:shell">

		Requiere Go 1.22 o superior.

		```bash
		git clone https://github.com/antojsh/velo-deploy.git
		cd velo-deploy
		go build -o velo-deploy ./cmd/velo-deploy
		sudo install -m 0755 velo-deploy /usr/local/bin/velo-deploy
		sudo ./install.sh
		```

	</TabItem>
</Tabs>

## Qué hace el instalador

1. Instala Caddy y lo habilita como servicio systemd.
2. Instala `nvm` en `/opt/nvm` para gestionar versiones de Node.js.
3. Copia `velo-deploy` a `/usr/local/bin/velo-deploy`.
4. Crea el directorio de configuración en `/etc/velo-deploy/`.
5. Crea el directorio de apps en `/opt/deploy/apps/`.
6. Crea el directorio de logs en `/var/log/velo-deploy/`.
7. Instala e inicia el servicio systemd `velo-deploy-watcher` en el puerto `9999`.

## Chequeos post-instalación

```bash
# El binario está en el PATH
which velo-deploy

# Caddy está corriendo
systemctl status caddy

# El watcher de webhooks está corriendo
systemctl status velo-deploy-watcher

# La config por defecto existe
cat /etc/velo-deploy/config.json
```

Si los cuatro chequeos pasan, estás listo para desplegar tu primera app.

## Siguientes pasos

- [Deploy rápido →](/velo-deploy/es/getting-started/quick-deploy/)
- [Walkthrough de la primera app →](/velo-deploy/es/getting-started/first-app/)
