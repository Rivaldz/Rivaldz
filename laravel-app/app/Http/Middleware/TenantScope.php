<?php

namespace App\Http\Middleware;

use App\Models\Merchant;
use Closure;
use Illuminate\Http\Request;
use Symfony\Component\HttpFoundation\Response;

class TenantScope
{
    /**
     * Resolve the tenant (merchant) belonging to the authenticated user and
     * bind it to the request so every downstream query can be scoped.
     *
     * Queries executed through the authenticated API guards are additionally
     * filtered by the global scope defined in App\Scopes\TenantScope.
     */
    public function handle(Request $request, Closure $next): Response
    {
        $user = $request->user('api');

        if ($user) {
            $merchant = $user->merchant;

            $request->attributes->set('tenant_merchant', $merchant);
            $request->attributes->set('tenant_merchant_id', $merchant?->id);

            if ($merchant) {
                app()->instance(Merchant::class, $merchant);
                app()->instance('tenant_merchant_id', $merchant->id);
            } else {
                app()->forgetInstance('tenant_merchant_id');
            }
        }

        return $next($request);
    }
}
