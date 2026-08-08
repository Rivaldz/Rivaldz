<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    /**
     * Performance indexes to speed up the daily revenue report queries.
     *
     * The report queries always filter by a tenant scope (merchant_id / outlet_id)
     * combined with a date range on created_at, then GROUP BY DATE(created_at).
     * A composite index that matches the query predicates lets MySQL use an
     * index range scan and avoids a full table scan.
     */
    public function up(): void
    {
        Schema::table('transactions', function (Blueprint $table) {
            $table->index(['merchant_id', 'created_at'], 'idx_transactions_merchant_created');
            $table->index(['outlet_id', 'created_at'], 'idx_transactions_outlet_created');
        });
    }

    /**
     * Reverse the migrations.
     */
    public function down(): void
    {
        Schema::table('transactions', function (Blueprint $table) {
            $table->dropIndex('idx_transactions_merchant_created');
            $table->dropIndex('idx_transactions_outlet_created');
        });
    }
};
