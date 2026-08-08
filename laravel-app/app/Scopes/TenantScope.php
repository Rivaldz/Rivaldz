<?php

namespace App\Scopes;

use Illuminate\Database\Eloquent\Builder;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Scope;

/**
 * Strict multi-tenancy global scope.
 *
 * Automatically restricts every Transaction / Outlet query to the merchant
 * bound for the current request by the TenantScope middleware. When no
 * tenant is bound (e.g. running in console/queue without an authenticated
 * user) the scope is a no-op so system-level commands still work.
 */
class TenantScope implements Scope
{
    public function apply(Builder $builder, Model $model): void
    {
        $merchantId = app()->bound('tenant_merchant_id') ? app('tenant_merchant_id') : null;

        if ($merchantId !== null) {
            $builder->where($model->qualifyColumn('merchant_id'), $merchantId);
        }
    }
}
