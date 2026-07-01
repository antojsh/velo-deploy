---
title: Velo Deploy
description: PaaS Bare Metal — Despliega aplicaciones Node.js y sitios estáticos en cualquier VPS sin Docker.
template: splash
hero:
  tagline: Un PaaS liviano y seguro que usa systemd para aislar procesos y Caddy para HTTPS automático. Sin contenedores, sin Kubernetes, sin overhead.
  actions:
    - text: Empezar
      link: getting-started/installation/
      icon: right-arrow
      variant: primary
    - text: Ver en GitHub
      link: https://github.com/antojsh/velo-deploy
      icon: external
      variant: minimal
  image:
    file: ../../../assets/logo-dark.svg
    alt: Logo de Velo Deploy
---

import { Card, CardGrid, LinkCard } from '@astrojs/starlight/components';

<CardGrid stagger>
	<Card title="Instalación en una línea" icon="rocket">
		De cero a funcionando en segundos. Un solo `curl` instala el daemon, la TUI, la CLI y todos los servicios necesarios.

		```bash
		curl -sS https://get.velo-deploy.sh | bash
		```
	</Card>
	<Card title="HTTPS automático" icon="seti:lock">
		Caddy emite y renueva certificados de Let's Encrypt para cada dominio. Sin gestión manual, sin downtime.
	</Card>
	<Card title="Node.js y sitios estáticos" icon="seti:javascript">
		Despliega servicios Node de larga duración y assets estáticos desde la misma CLI. Detecta automáticamente el directorio de build.
	</Card>
	<Card title="Aislamiento con systemd" icon="seti:default">
		Cada app corre como un usuario Linux propio con una unidad systemd endurecida: sistema read-only, tmp aislado, sin escalada de privilegios.
	</Card>
	<Card title="Webhooks de GitHub" icon="github">
		Push a `main` y Velo hace pull, build y restart. Sin pipeline de CI.
	</Card>
	<Card title="Dashboard TUI" icon="seti:terminal">
		Gestioná todas tus apps desde una sola interfaz de terminal manejada con teclado. Sin UI web que mantener.
	</Card>
</CardGrid>

## ¿Por qué Velo Deploy?

Las plataformas basadas en contenedores son poderosas, pero la mayoría de los equipos solo necesitan un subconjunto chico de sus features. Velo Deploy trae lo justo para desplegar una app Node.js o un sitio estático en un único VPS, con las garantías de seguridad que esperarías de un PaaS.

<LinkCard
	title="Leer la arquitectura →"
	description="Entendé cómo encajan systemd, Caddy y el binario de Go."
	href="architecture/overview/"
/>

## Navegación rápida

<CardGrid>
	<LinkCard title="Instalación" href="getting-started/installation/" description="Requisitos del sistema e instalador de una línea." />
	<LinkCard title="Deploy rápido" href="getting-started/quick-deploy/" description="De un URL de Git a una app corriendo en 30 segundos." />
	<LinkCard title="Referencia de CLI" href="reference/cli-reference/" description="Cada comando, flag y código de salida." />
	<LinkCard title="Troubleshooting" href="operations/troubleshooting/" description="Diagnosticá los problemas más comunes en producción." />
</CardGrid>
