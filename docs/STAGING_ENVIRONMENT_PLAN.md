# Staging environment — plan

**Status:** ⚠️ SUPERSEDED. Written as a plan; staging is now live and has been
since 2026-09-07. Kept for the reasoning in §2 (why a port cannot simply be
opened). For what is actually running, see [`QUOTATION_PROGRESS.md`](QUOTATION_PROGRESS.md) §1.
**Date:** 2026-09-07
**Goal:** deploy and test a release on the same host before it reaches production.

---

## 1. What is on the server today

Host `108.136.218.247`, SSH on port 4722, Ubuntu 24.04, uptime 28 weeks.

| Container | Image | Bound to |
|---|---|---|
| `frontend` | `backent/tmn-mapping-frontend:2.45.0` | `127.0.0.1:3000` → 80 |
| `backend` | `backent/tmn-mapping-backend:2.30.0` | `127.0.0.1:8080` → 8088 |
| `tmn-grafana` | `grafana/grafana:10.1.0` | `127.0.0.1:33000` |
| `tmn-loki` | `grafana/loki:3.3.2` | `0.0.0.0:3100` |
| `tmn-promtail` | `grafana/promtail:3.3.2` | — |

- **Database is AWS RDS**, not on the box: `tmn-iris-db…ap-southeast-3.rds.amazonaws.com`,
  database `tmn_mapping`. Note the local dev database is named `tmn_backend`.
- **nginx on the host** terminates 443 with a Certbot certificate for `tmnmapping.com`
  and proxies `/api/` → `127.0.0.1:8080`, `/` → `127.0.0.1:3000`.
- Docker network `global-network` joins the app containers.
- Backend config comes from `/home/ubuntu/.env` (19 keys, including RDS credentials,
  `APP_SECRET_KEY` and the ERP API credentials).
- Deploys come from Jenkins, which SSHes in and runs:
  ```
  sudo docker run -dp 127.0.0.1:8080:8088 \
      --env-file .env --network global-network \
      --name backend --restart unless-stopped \
      backent/tmn-mapping-backend:<version>
  ```
- Resources: **7.8 GB RAM (6.7 GB available)**, **48 GB disk at 74% — 13 GB free**.
- **No host firewall.** `ufw` is inactive, so inbound access is governed entirely by
  the AWS Security Group.

---

## 2. ⚠️ Can we expose a port? Not without an AWS change

**Verified:** `tmn-loki` binds `0.0.0.0:3100`, yet connecting to
`http://108.136.218.247:3100/ready` from outside **times out**, while ports 80 and 443
both answer. That proves the AWS Security Group permits only **80, 443 and 4722**.

Binding a staging container to `0.0.0.0:<port>` therefore achieves nothing on its own —
the packet never reaches the host. Opening a port requires an inbound rule on the
instance's Security Group, which lives in the AWS console/API, **not on the server**.
It cannot be done over SSH and is outside what I can reach.

### Four ways to reach staging

| Option | Needs | Reachable by | Verdict |
|---|---|---|---|
| **A. SSH tunnel** | Nothing. Works today | Anyone with the `.pem` | **Start here.** Zero infrastructure change |
| **B. Open a port in the Security Group** | You add an inbound rule in AWS | Anyone (or a restricted CIDR) | Best for team access. Restrict to your office/VPN CIDR, not `0.0.0.0/0` |
| **C. Subdomain** `staging.tmnmapping.com` | DNS A record + `certbot --nginx -d staging.tmnmapping.com` | Anyone, over HTTPS | Cleanest long-term, but it is a domain — which you wanted to avoid |
| **D. Path prefix** `tmnmapping.com/staging/` | nginx `location` block | Anyone | Avoid. The Vue SPA needs rebuilding with a base path, and the router base changes |

**Option A, available immediately once staging is up:**

```bash
ssh -p 4722 -i tmn-app-key.pem \
    -L 3001:127.0.0.1:3001 \
    -L 8081:127.0.0.1:8081 \
    ubuntu@108.136.218.247
```

Then open `http://localhost:3001`. The tunnel maps staging's ports onto your machine;
nothing is exposed publicly.

**Option B, once you have added the rule** — say TCP 8090 from your CIDR — add an
nginx server block listening on 8090 that proxies to the staging containers, so
staging keeps the same `/api/` path split as production. Binding Docker straight to
`0.0.0.0:8090` also works but skips nginx, and then the frontend's `/api` calls have
nowhere to go.

---

## 3. The database is the decision that matters

Staging **must not** share `tmn_mapping`. Testing migration 015's role backfill against
production data is exactly the scenario where a mistake is unrecoverable.

| Option | Isolation | Cost | Risk |
|---|---|---|---|
| **New database on the same RDS instance** (`tmn_mapping_staging`) | Logical only | None extra | A typo in `.env.staging` points staging at **production**. Same host, same credentials |
| **Postgres + PostGIS container on the box** | Complete | ~600 MB image + data | None to production. Diverges from RDS in SSL and connection handling |

**Recommendation: the container.** The blast radius argument outweighs fidelity. A bad
migration on staging then costs nothing, and `docker rm -f` resets it. The connection
differences (`POSTGRES_SSLMODE`) are a config line, not a code path.

Seed it from a production dump so the Phase 0 backfill is tested against real role
values:

```bash
# read-only against production
pg_dump --no-owner --no-acl -h <rds-host> -U <user> -d tmn_mapping > /tmp/prod.sql
# restore into the staging container
psql -h 127.0.0.1 -p 5433 -U postgres -d tmn_mapping_staging < /tmp/prod.sql
```

---

## 4. Proposed layout

| Piece | Production | Staging |
|---|---|---|
| Backend | `backend` · `127.0.0.1:8080` | `backend-staging` · `127.0.0.1:8081` |
| Frontend | `frontend` · `127.0.0.1:3000` | `frontend-staging` · `127.0.0.1:3001` |
| Database | RDS `tmn_mapping` | container `postgres-staging` · `127.0.0.1:5433` |
| Config | `/home/ubuntu/.env` | `/home/ubuntu/.env.staging` |
| Network | `global-network` | `staging-network` (separate, so a misconfigured host name cannot resolve to a production container) |
| Access | `https://tmnmapping.com` | SSH tunnel, then a Security Group port |

`.env.staging` differs from `.env` in: `POSTGRES_HOST=postgres-staging`,
`POSTGRES_PORT=5432`, `POSTGRES_DATABASE=tmn_mapping_staging`, `POSTGRES_SSLMODE=disable`,
its own `APP_SECRET_KEY` (so a staging token is never valid in production), and
`ENVIRONMENT=staging`.

### The ERP sync needs a decision

`main.go` runs a **full ERP sync on startup** and every `ERP_SYNC_INTERVAL_MINUTES`
after. A staging backend will therefore hit the real ERP API as well, doubling the
load — the last production cycle pulled 3,731 buildings.

It only *writes* to the staging database, so nothing is corrupted. But set
`ERP_SYNC_INTERVAL_MINUTES=1440` in `.env.staging` to cut the recurring load. The
startup sync cannot be disabled by config today; if that matters, it needs a
`ERP_SYNC_ON_STARTUP` flag in code.

---

## 5. Deploy path

The Jenkinsfile already supports `DEPLOY_TYPE=VERSION_ONLY` with a `TARGET_VERSION`,
which is most of what staging needs. Two options:

1. **Parameterise the existing pipeline** — add a `TARGET_ENV` parameter that switches
   container name, port and env file. Smallest change; one pipeline to maintain.
2. **A second Jenkins job** pointing at the same repo with staging defaults. Clearer
   separation, more duplication.

Either way the intended flow becomes: build → push → **deploy to staging** → verify →
promote the *same image tag* to production with `VERSION_ONLY`. Promoting the tag,
rather than rebuilding, is what makes the test meaningful.

---

## 6. Before starting

1. **Prune first.** Disk is at 74% with 13 GB free; Docker reports 62 images and
   ~620 MB reclaimable. A staging stack adds roughly 1.5 GB.
2. **Decide the database option** (§3).
3. **Decide the access route** (§2) — tunnel now, Security Group when the team needs it.
4. **Confirm production's migration version.** Still unknown: my read attempt failed
   authentication, and I stopped rather than retry credentials against production RDS.
   This matters — migrations 015 and 016 must run on staging first, and we do not yet
   know how far behind production is.
5. `ENVIRONMENT=development` is set on the **production** backend. Unrelated to
   staging, but worth correcting while nearby.

Nothing above has been executed. Every command run against the server so far has been
read-only, with one exception: pulling `postgres:17-alpine` (~150 MB), which is still
on the box and can be removed.
