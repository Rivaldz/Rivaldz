<?php

use App\Http\Controllers\Api\AuthController;
use App\Http\Controllers\Api\ReportController;
use Illuminate\Support\Facades\Route;

Route::prefix('auth')->group(function () {
    Route::post('login', [AuthController::class, 'login'])->name('auth.login');
    Route::post('register', [AuthController::class, 'register'])->name('auth.register');

    Route::middleware('auth:api')->group(function () {
        Route::post('logout', [AuthController::class, 'logout'])->name('auth.logout');
        Route::post('refresh', [AuthController::class, 'refresh'])->name('auth.refresh');
        Route::get('me', [AuthController::class, 'me'])->name('auth.me');
    });
});

Route::prefix('reports')
    ->middleware(['auth:api', 'tenant.scope'])
    ->group(function () {
        Route::get('merchant/daily', [ReportController::class, 'merchantDaily'])
            ->name('reports.merchant.daily');

        Route::get('outlet/daily', [ReportController::class, 'outletDaily'])
            ->name('reports.outlet.daily');
    });
