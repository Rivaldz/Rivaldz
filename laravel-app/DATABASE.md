# Database Documentation — Majoo Revenue Reporting

## Overview

The application uses a MySQL schema (InnoDB) with the three business tables defined in
the test brief, plus a Laravel-standard `users` table used for JWT authentication.

| Table          | Purpose                                              | Tenant key      |
|----------------|------------------------------------------------------|-----------------|
| `users`        | Login accounts (`merchants.user_id` → `users.id`)    | –               |
| `merchants`    | Merchant tenants, owned by one user                  | `id`            |
| `outlets`      | Outlets, each owned by exactly one merchant          | `merchant_id`   |
| `transactions` | Sales transactions (each has `bill_total`)           | `merchant_id`, `outlet_id` |

> Local development / automated tests in this repo run on **SQLite** (see `.env`).
> The migrations are written with the Laravel Schema Builder so they are portable;
> the SQL below is the MySQL form used in production.

---

## 1. DML Documentation

All data access in the application goes through Eloquent models. The statements below
are the actual SQL generated for each feature.

### 1.1 Authentication (JWT)

**Find the user during login** (Eloquent `Auth::guard('api')->attempt()`):

```sql
SELECT * FROM `users`
WHERE `email` = 'merchant1@example.com'
LIMIT 1;
```

**Resolve tenant after login** (middleware `App\Http\Middleware\TenantScope`,
`User::merchant()`):

```sql
SELECT * FROM `merchants`
WHERE `merchants`.`user_id` = 1
LIMIT 1;
```

### 1.2 Merchant daily revenue report (`GET /api/reports/merchant/daily`)

The controller builds the full month calendar (e.g. 2026-11-01 … 2026-11-30) in PHP,
then runs a single grouped aggregation. The `merchant_id` predicate appears twice as a
defence-in-depth measure (explicit controller scope **and** the global tenant scope
`App\Scopes\TenantScope`). Days missing from the result set are reported as `0`.

```sql
SELECT DATE(created_at) AS date,
       SUM(bill_total)  AS revenue
FROM `transactions`
WHERE `created_at` BETWEEN '2026-11-01 00:00:00' AND '2026-11-30 23:59:59'
  AND `merchant_id` = 1                -- explicit tenant scope (controller)
  AND `transactions`.`merchant_id` = 1 -- global tenant scope (model)
GROUP BY DATE(created_at);
```

### 1.3 Outlet daily revenue report (`GET /api/reports/outlet/daily`)

**Tenant check on the requested outlet** (returns 403 when not owned by the caller):

```sql
SELECT * FROM `outlets`
WHERE `id` = 1
LIMIT 1;
```

**Daily revenue for that outlet:**

```sql
SELECT DATE(created_at) AS date,
       SUM(bill_total)  AS revenue
FROM `transactions`
WHERE `created_at` BETWEEN '2026-08-01 00:00:00' AND '2026-08-31 23:59:59'
  AND `outlet_id` = 1
  AND `transactions`.`merchant_id` = 1 -- global tenant scope (model)
GROUP BY DATE(created_at);
```

### 1.4 `GET /api/auth/me` (JWT context)

```sql
SELECT * FROM `users` WHERE `id` = 1 LIMIT 1;
SELECT * FROM `merchants` WHERE `merchants`.`user_id` = 1 LIMIT 1;
```

### 1.5 Writes used by the seeder

All inserts in `DatabaseSeeder` go through `updateOrCreate` (a `SELECT` then an
`INSERT`/`UPDATE`), so they are idempotent. Example row:

```sql
INSERT INTO `transactions`
  (`merchant_id`, `outlet_id`, `bill_total`, `created_at`, `created_by`,
   `updated_at`, `updated_by`)
VALUES
  (1, 1, 2000, '2026-08-01 12:30:04', 1, '2026-08-01 12:30:04', 1);
```

---

## 2. Indexing Strategy

### 2.1 Index inventory

| Index                       | Table          | Columns                     | Created by                         |
|-----------------------------|----------------|-----------------------------|------------------------------------|
| `PRIMARY`                   | `transactions` | `id`                        | schema                             |
| `idx_transactions_merchant_id` | `transactions` | `merchant_id`               | migration `…000003`              |
| `idx_transactions_outlet_id`   | `transactions` | `outlet_id`                 | migration `…000003`              |
| `idx_transactions_merchant_created` | `transactions` | `merchant_id`, `created_at` | migration `…000004` (composite)    |
| `idx_transactions_outlet_created`   | `transactions` | `outlet_id`, `created_at`   | migration `…000004` (composite)    |
| `idx_outlets_merchant_id`   | `outlets`      | `merchant_id`               | migration `…000002`              |
| `idx_merchants_user_id`     | `merchants`    | `user_id`                   | migration `…000001`              |

### 2.2 Why composite indexes on `transactions`

Every report query has the same shape:

```sql
WHERE <tenant_key> = ? AND created_at BETWEEN ? AND ?
GROUP BY DATE(created_at);
```

Without a suitable index MySQL must perform a **full table scan**, and for the
aggregation the entire filtered row set is read to compute `SUM(bill_total)`. With a
composite index `(merchant_id, created_at)`:

1. **Index range scan**: the equality predicate on `merchant_id` narrows the index to a
   single B+Tree subtree; `created_at` (`BETWEEN`) then scans only the relevant index
   pages — the query touches just the tenant's rows for the month instead of all rows.
2. **Covering-ish access**: the index stores both predicate columns; MySQL reads the
   index pages to locate the rows before fetching the table row for `bill_total`.
3. **Ordered output for `GROUP BY DATE(...)`**: an index ordered by
   `(merchant_id, created_at)` yields date values in ascending order, letting the
   optimizer avoid a filesort/temporary table in the common case.

The same reasoning applies to `(outlet_id, created_at)` for the outlet report.

> **Leftmost-prefix rule**: this is why the composite index is ordered
> `tenant_key` **then** `created_at`. An index on `(created_at, merchant_id)` would
> still require a scan over all dates before the equality filter, defeating the purpose.

### 2.3 Lookup indexes for tenant resolution

- `merchants(user_id)` — the `TenantScope` middleware resolves the merchant via
  `WHERE user_id = ?` on every authenticated request. A unique-ish index turns this
  into a point lookup instead of a scan.
- `outlets(merchant_id)` — supports the tenant check `WHERE id = ?` joined with the
  merchant scope, and any future "outlets of a merchant" listing.

### 2.4 `EXPLAIN` verification

Run against MySQL to confirm index usage:

```sql
EXPLAIN
SELECT DATE(created_at) AS date, SUM(bill_total) AS revenue
FROM transactions
WHERE merchant_id = 1
  AND created_at BETWEEN '2026-08-01 00:00:00' AND '2026-08-31 23:59:59'
GROUP BY DATE(created_at);
```

Expected `key` = `idx_transactions_merchant_created` with `type` = `range`
(ref `ref`, key parts `merchant_id, created_at`).

---

## 3. Multi-Tenancy Enforcement (data security)

1. **JWT layer**: every `api/reports/*` route is behind `auth:api` (JWT guard).
2. **Tenant resolution**: `TenantScope` middleware derives `merchant_id` from the
   authenticated user's `merchants` row — never from client input.
3. **Global scope**: `App\Scopes\TenantScope` is registered on the `Transaction` and
   `Outlet` models, automatically appending `merchant_id = <tenant>` to any Eloquent
   query in a request context.
4. **Explicit ownership check**: the outlet report verifies the requested `outlet_id`
   belongs to the caller's merchant and returns **403** otherwise.

---

## 4. Production MySQL setup

Install `pdo_mysql` (this dev environment only has `pdo_sqlite`) and set:

```dotenv
DB_CONNECTION=mysql
DB_HOST=127.0.0.1
DB_PORT=3306
DB_DATABASE=majoo
DB_USERNAME=root
DB_PASSWORD=secret
```

Then run `php artisan migrate --seed` to build the schema and load the sample data.
