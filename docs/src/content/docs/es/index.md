---
title: Velo Deploy
description: Despliega apps Node.js y sitios estáticos en cualquier VPS con aislamiento systemd, HTTPS automático y sin Docker.
template: splash
hero:
  title: Despliega Node.js como un servicio Linux serio
  tagline: >-
    Convierte un VPS común en una pequeña plataforma pensada para producción:
    un binario Go, HTTPS con Caddy, unidades systemd, aislamiento por usuario
    Linux, deploys por webhook y ningún daemon Docker en el camino crítico.
  actions:
    - text: Deploy en 30 segundos
      link: getting-started/quick-deploy/
      icon: right-arrow
      variant: primary
    - text: Entender la arquitectura
      link: architecture/overview/
      icon: external
      variant: minimal
---

<div class="velo-pill-row">
  <span class="velo-pill">Sin daemon Docker</span>
  <span class="velo-pill">Deploys nativos con systemd</span>
  <span class="velo-pill">HTTPS con Caddy</span>
  <span class="velo-pill">Aislamiento por usuario Linux</span>
  <span class="velo-pill">Push-to-deploy</span>
</div>

<div class="velo-metric-row">
  <span class="velo-metric"><strong>1</strong> binario Go</span>
  <span class="velo-metric"><strong>0</strong> control plane en runtime</span>
  <span class="velo-metric"><strong>24</strong> Node.js LTS por defecto</span>
</div>

<span class="velo-section-kicker">Base de producción</span>

## La capa de deploy para VPS de quienes entienden Linux

No necesitas un barco de contenedores para mover una bicicleta. Para muchas APIs Node.js, builds de Astro, dashboards, paneles admin y sitios estáticos, el target más limpio sigue siendo Linux: un proceso supervisado por `systemd`, detrás de Caddy y corriendo con su propio usuario.

Velo Deploy empaqueta esa base aburrida, potente y confiable en un flujo que un equipo puede operar de verdad.

<div class="velo-command-panel">
  <div class="velo-command-header">
    <span class="velo-dot"></span>
    <span class="velo-dot"></span>
    <span class="velo-dot"></span>
    <span>vps-limpio → app HTTPS en vivo</span>
  </div>

```bash
curl -sS https://github.com/antojsh/velo-deploy/releases/latest/download/velo-deploy-install.sh | bash
velo-deploy deploy https://github.com/tu/app-node-o-sitio-estatico
git push origin main
```

</div>

<span class="velo-section-kicker">Elige tu camino</span>

## Empieza con el modelo mental correcto

<div class="velo-doc-grid">
  <a class="velo-doc-card" href="getting-started/installation/">
    <span>01 · Preparar</span>
    <strong>Instala una vez en el VPS</strong>
    <p>Revisa requisitos, instala el binario y verifica Caddy junto al watcher de webhooks.</p>
  </a>
  <a class="velo-doc-card" href="getting-started/quick-deploy/">
    <span>02 · Desplegar</span>
    <strong>Publica un repo rápido</strong>
    <p>Pasa de una URL de Git a una app Node.js o estática corriendo con HTTPS.</p>
  </a>
  <a class="velo-doc-card" href="architecture/security/">
    <span>03 · Confiar</span>
    <strong>Entiende el aislamiento</strong>
    <p>Mira cómo usuarios Linux, permisos y unidades systemd endurecidas contienen cada app.</p>
  </a>
</div>

<span class="velo-section-kicker">Por qué existe Velo</span>

## Una plataforma más pequeña, hecha con piezas aburridas

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
    <span class="velo-icon run">04</span>
    <strong>Node.js y sitios estáticos</strong>
    <p>Despliega servicios Node.js de larga vida o directorios estáticos como <code>dist</code>, <code>build</code> y <code>public</code>.</p>
  </div>
  <div class="velo-card">
    <span class="velo-icon run">05</span>
    <strong>Push-to-deploy</strong>
    <p>El watcher de webhooks puede redesplegar con cada push. Flujo Git simple, sin depender de una plataforma externa.</p>
  </div>
  <div class="velo-card">
    <span class="velo-icon run">06</span>
    <strong>Un binario Go pequeño</strong>
    <p>Sin control plane en runtime. Velo agrega menos peso operativo que un proceso Node.js promedio.</p>
  </div>
</div>

<span class="velo-section-kicker">Límites arquitectónicos</span>

## PM2 no es una arquitectura de despliegue

<div class="velo-quote">
PM2 sirve, pero poner un process manager de Node.js a cuidar otros procesos Node.js es como poner un ladrón a cuidar a otro ladrón. A veces funciona. No es el cimiento que quieres para una plataforma en VPS.
</div>

<div class="velo-compare">
  <div class="velo-compare-card bad">
    <span class="velo-icon warn">!</span>
    <strong>Deploys centrados en PM2</strong>
    <p>Buenos para empezar rápido, débiles como modelo operativo. Todavía debes resolver usuarios Linux, HTTPS, arranque al boot, logs, reinicios, límites de recursos y seguridad del host.</p>
  </div>
  <div class="velo-compare-card good">
    <span class="velo-icon run">✓</span>
    <strong>Velo Deploy</strong>
    <p>Usa el sistema operativo como supervisor: systemd para ciclo de vida, Caddy para HTTPS, usuarios Linux para aislamiento y una CLI en Go para conectar las piezas.</p>
  </div>
</div>

## Docker es poderoso. No siempre es necesario.

Docker brilla cuando necesitas portabilidad de imágenes, entornos multi-servicio u orquestación. Pero si tu target es un VPS y tu workload es una app Node.js o un sitio estático, los contenedores pueden convertirse en complejidad accidental.

<div class="velo-principles">
  <div class="velo-principle">
    <strong>Usa el sistema operativo como plataforma</strong>
    <p><code>systemd</code> supervisa procesos, recupera al boot y expone logs con herramientas estándar de Linux.</p>
  </div>
  <div class="velo-principle">
    <strong>Mantén la entrega con forma de Git</strong>
    <p>Despliega desde un repositorio y, si quieres, deja que los webhooks redesplieguen cuando cambie <code>main</code>.</p>
  </div>
  <div class="velo-principle">
    <strong>Haz que HTTPS sea aburrido</strong>
    <p>Caddy se encarga del reverse proxy y la renovación de certificados, para que la app piense solo en código.</p>
  </div>
  <div class="velo-principle">
    <strong>Prefiere versiones explícitas</strong>
    <p>Node 24 LTS es el default; <code>engines.node</code> permite fijar otro major soportado por app.</p>
  </div>
</div>

<div class="velo-footer-cta">

## Empieza donde está el valor

<div class="velo-roadmap">
  <a href="getting-started/installation/"><span class="velo-step-number">1</span><span><b>Instalar Velo Deploy</b><br><small>Prepara el VPS una sola vez.</small></span></a>
  <a href="getting-started/quick-deploy/"><span class="velo-step-number">2</span><span><b>Deploy rápido en 30 segundos</b><br><small>Despliega un repo ahora.</small></span></a>
  <a href="architecture/security/"><span class="velo-step-number">3</span><span><b>Modelo de seguridad</b><br><small>Entiende cómo funciona el aislamiento.</small></span></a>
  <a href="guide/node-versions/"><span class="velo-step-number">4</span><span><b>Versiones de Node.js</b><br><small>Usa el default o fija la versión de tu app.</small></span></a>
</div>

</div>
