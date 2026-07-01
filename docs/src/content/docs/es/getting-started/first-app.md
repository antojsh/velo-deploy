---
title: Primera app
description: Construí y desplegá una API Node.js real desde cero.
---

import { Steps, Aside } from '@astrojs/starlight/components';

Este walkthrough despliega una API mínima de Node.js en tu VPS y la expone bajo un alias local.

## 1. Creá el proyecto local

<Steps>

1. Creá una carpeta nueva e inicializá el proyecto Node.

	```bash
	mkdir hello-velo && cd hello-velo
	npm init -y
	```

2. Agregá un servidor HTTP mínimo. Creá `index.js`:

	```js
	import { createServer } from 'node:http';

	const port = process.env.PORT || 3000;

	createServer((req, res) => {
	  res.writeHead(200, { 'content-type': 'application/json' });
	  res.end(JSON.stringify({ hello: 'velo', path: req.url }));
	}).listen(port, () => {
	  console.log(`escuchando en ${port}`);
	});
	```

3. Pinea la versión de Node con `engines` y agregá un script de start.

	```json
	{
	  "name": "hello-velo",
	  "type": "module",
	  "engines": { "node": ">=20" },
	  "scripts": { "start": "node index.js" }
	}
	```

4. Pusheá el proyecto a un repo de Git (GitHub, Gitea, self-hosted — cualquier cosa que Velo pueda clonar con `git`).

	```bash
	git init
	git add .
	git commit -m "feat: hello world mínimo"
	gh repo create hello-velo --public --source=. --push
	```

</Steps>

## 2. Desplegar desde el VPS

Conectate al VPS por SSH y ejecutá:

```bash
velo-deploy deploy https://github.com/tu-usuario/hello-velo
```

Velo clona, instala (no hay dependencias en este caso, pero va a correr `npm install` si las tenés), elige Node 20 desde el campo `engines`, registra una unidad systemd y escribe un vhost de Caddy.

## 3. Verificar

```bash
velo-deploy list
```

Deberías ver `hello-velo` con un alias local y un puerto asignado. Agregá el alias al `/etc/hosts` de tu laptop y hacé curl:

```bash
curl http://hello-velo.local
# {"hello":"velo","path":"/"}
```

## 4. Agregar un dominio (opcional)

Cuando estés listo para compartir la API:

```bash
velo-deploy add hello-velo /opt/deploy/apps/hello-velo --domain api.example.com
```

Caddy va a pedir un certificado de Let's Encrypt la primera vez que el dominio resuelva a tu servidor. Mirá [Dominios custom](/velo-deploy/es/guide/domains/) para los pasos de DNS.

<Aside type="tip" title="Mirá los logs mientras iterás">
	Abrí una segunda sesión de SSH y tail de los logs:
	```bash
	velo-deploy logs hello-velo -f
	```
	Cada restart desde `velo-deploy restart` o un deploy por webhook aparece acá.
</Aside>

## 5. Activar auto-deploy

En el servidor:

```bash
sudo systemctl enable --now velo-watcher
```

En tu repo de GitHub, agregá un webhook con:

- **Payload URL**: `http://tu-servidor:9999/webhook`
- **Content type**: `application/json`
- **Events**: Solo el evento push

Hacé push a `main` y Velo va a hacer pull, rebuild y restart automáticamente. Mirá [Webhooks](/velo-deploy/es/guide/webhooks/) para la configuración completa.

## Siguientes pasos

- [Referencia de CLI →](/velo-deploy/es/reference/cli-reference/)
- [Arquitectura →](/velo-deploy/es/architecture/overview/)
