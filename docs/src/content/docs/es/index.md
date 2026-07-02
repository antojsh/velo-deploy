---
title: Velo Deploy
description: Desplegá apps Node.js y sitios estáticos en cualquier VPS con aislamiento systemd, HTTPS automático y sin Docker.
template: splash
hero:
  title: Desplegá Node.js como un servicio Linux serio
  tagline: >-
    Velo Deploy convierte un VPS común en una mini plataforma para apps Node.js
    y sitios estáticos: un binario Go, HTTPS con Caddy, unidades systemd,
    usuarios Linux y cero impuesto Docker.
  actions:
    - text: Deploy en 30 segundos
      link: getting-started/quick-deploy/
      icon: right-arrow
      variant: primary
    - text: Ver arquitectura
      link: architecture/overview/
      icon: external
      variant: minimal
---

<div class="velo-pill-row">
  <span class="velo-pill">Sin daemon Docker</span>
  <span class="velo-pill">Sin ceremonia Kubernetes</span>
  <span class="velo-pill">Deploys nativos con systemd</span>
  <span class="velo-pill">HTTPS automático</span>
  <span class="velo-pill">Aislamiento por usuario Linux</span>
</div>

## La capa de deploy para VPS de quienes entienden producción

No necesitas un barco de contenedores para mover una bicicleta. Para muchas APIs Node.js, builds de Astro, dashboards, paneles admin y sitios estáticos, el target más limpio sigue siendo Linux: un proceso supervisado por `systemd`, detrás de Caddy y corriendo con su propio usuario.

Velo Deploy empaqueta esa base aburrida, potente y confiable en un flujo que un equipo puede usar sin pelearse con la infraestructura.

<div class="velo-command-panel">
  <div class="velo-command-header">
    <span class="velo-dot"></span>
    <span class="velo-dot"></span>
    <span class="velo-dot"></span>
    <span>vps-limpio → app HTTPS en vivo</span>
  </div>

<pre><code class="language-bash">curl -sS https://github.com/antojsh/velo-deploy/releases/latest/download/velo-deploy-install.sh | bash
velo-deploy deploy https://github.com/tu/app-node-o-sitio-estatico
git push origin main</code></pre>

</div>

## ¿Por qué desplegar con Velo?

<div class="velo-grid">
  <div class="velo-card">
    <span class="velo-icon">01</span>
    <strong>Sin impuesto Docker</strong>
    <p>Evita daemon, builds de imágenes, registries, compose files y redes de contenedores cuando tu app solo necesita un servicio Linux confiable.</p>
  </div>
  <div class="velo-card">
    <span class="velo-icon">02</span>
    <strong>Aislamiento seguro en Linux</strong>
    <p>Cada app corre como su propio usuario Linux con hardening de systemd: <code>ProtectSystem</code>, <code>PrivateTmp</code> y <code>NoNewPrivileges</code>.</p>
  </div>
  <div class="velo-card">
    <span class="velo-icon">03</span>
    <strong>HTTPS automático</strong>
    <p>Caddy gestiona certificados y renovaciones. Apunta el DNS al VPS, despliega la app y deja de cuidar TLS a mano.</p>
  </div>
  <div class="velo-card">
    <span class="velo-icon">04</span>
    <strong>Node.js y sitios estáticos</strong>
    <p>Despliega servicios Node.js de larga vida o directorios estáticos como <code>dist</code>, <code>build</code> y <code>public</code>.</p>
  </div>
  <div class="velo-card">
    <span class="velo-icon">05</span>
    <strong>Push-to-deploy</strong>
    <p>El watcher de webhooks puede redesplegar con cada push. Flujo Git simple, sin depender de una plataforma externa.</p>
  </div>
  <div class="velo-card">
    <span class="velo-icon">06</span>
    <strong>Un binario Go pequeño</strong>
    <p>Sin control plane en runtime. Velo agrega menos peso operativo que un proceso Node.js promedio.</p>
  </div>
</div>

## PM2 no es una arquitectura de despliegue

<div class="velo-quote">
PM2 sirve, pero poner un process manager de Node.js a cuidar otros procesos Node.js es como poner un ladrón a cuidar a otro ladrón. A veces funciona. No es el cimiento que quieres para una plataforma en VPS.
</div>

<div class="velo-compare">
  <div class="velo-compare-card bad">
    <strong>Deploys centrados en PM2</strong>
    <p>Buenos para empezar rápido, débiles como modelo operativo. Todavía debes resolver usuarios Linux, HTTPS, arranque al boot, logs, reinicios, límites de recursos y seguridad del host.</p>
  </div>
  <div class="velo-compare-card good">
    <strong>Velo Deploy</strong>
    <p>Usa el sistema operativo como supervisor: systemd para ciclo de vida, Caddy para HTTPS, usuarios Linux para aislamiento y una CLI en Go para conectar las piezas.</p>
  </div>
</div>

## Docker es poderoso. No siempre es necesario.

Docker brilla cuando necesitas portabilidad de imágenes, entornos multi-servicio u orquestación. Pero si tu target es un VPS y tu workload es una app Node.js o un sitio estático, los contenedores pueden convertirse en complejidad accidental.

Velo elige primero las piezas simples:

- `systemd` para supervisión de procesos y recuperación al boot.
- Usuarios y permisos Linux para aislamiento.
- Caddy para reverse proxy y TLS.
- Git para entrega.
- Node 24 LTS por defecto, con overrides explícitos en `engines.node` cuando tu app necesita otra versión major soportada.

## Empieza donde está el valor

- [Instalar Velo Deploy](getting-started/installation/) — prepara el VPS una sola vez.
- [Deploy rápido en 30 segundos](getting-started/quick-deploy/) — despliega un repo ahora.
- [Modelo de seguridad](architecture/security/) — entiende cómo funciona el aislamiento.
- [Versiones de Node.js](guide/node-versions/) — usa el default o fija la versión de tu app.
