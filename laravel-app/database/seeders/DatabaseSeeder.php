<?php

namespace Database\Seeders;

use App\Models\Merchant;
use App\Models\Outlet;
use App\Models\Transaction;
use App\Models\User;
use Illuminate\Database\Seeder;

class DatabaseSeeder extends Seeder
{
    /**
     * Seed the application's database with the sample data from the
     * provided SQL schema (2 users, 2 merchants, 3 outlets, 22 transactions).
     */
    public function run(): void
    {
        $user1 = User::updateOrCreate(
            ['email' => 'merchant1@example.com'],
            ['name' => 'Merchant One', 'password' => 'password']
        );

        $user2 = User::updateOrCreate(
            ['email' => 'merchant2@example.com'],
            ['name' => 'Merchant Two', 'password' => 'password']
        );

        $merchant1 = Merchant::updateOrCreate(
            ['id' => 1],
            ['user_id' => $user1->id, 'merchant_name' => 'merchant 1', 'created_by' => $user1->id, 'updated_by' => $user1->id]
        );

        $merchant2 = Merchant::updateOrCreate(
            ['id' => 2],
            ['user_id' => $user2->id, 'merchant_name' => 'Merchant 2', 'created_by' => $user2->id, 'updated_by' => $user2->id]
        );

        $outlet1 = Outlet::updateOrCreate(
            ['id' => 1],
            ['merchant_id' => $merchant1->id, 'outlet_name' => 'Outlet 1', 'created_by' => $user1->id, 'updated_by' => $user1->id]
        );

        $outlet2 = Outlet::updateOrCreate(
            ['id' => 2],
            ['merchant_id' => $merchant2->id, 'outlet_name' => 'Outlet 1', 'created_by' => $user2->id, 'updated_by' => $user2->id]
        );

        $outlet3 = Outlet::updateOrCreate(
            ['id' => 3],
            ['merchant_id' => $merchant1->id, 'outlet_name' => 'Outlet 2', 'created_by' => $user1->id, 'updated_by' => $user1->id]
        );

        $transactions = [
            [1, 1, 1, 2000, '2026-08-01 12:30:04'],
            [2, 1, 1, 2500, '2026-08-01 17:20:14'],
            [3, 1, 1, 4000, '2026-08-02 12:30:04'],
            [4, 1, 1, 1000, '2026-08-04 12:30:04'],
            [5, 1, 1, 7000, '2026-08-05 16:59:30'],
            [6, 1, 3, 2000, '2026-08-02 18:30:04'],
            [7, 1, 3, 2500, '2026-08-03 17:20:14'],
            [8, 1, 3, 4000, '2026-08-04 12:30:04'],
            [9, 1, 3, 1000, '2026-08-04 12:31:04'],
            [10, 1, 3, 7000, '2026-08-05 16:59:30'],
            [11, 2, 2, 2000, '2026-08-01 18:30:04'],
            [12, 2, 2, 2500, '2026-08-02 17:20:14'],
            [13, 2, 2, 4000, '2026-08-03 12:30:04'],
            [14, 2, 2, 1000, '2026-08-04 12:31:04'],
            [15, 2, 2, 7000, '2026-08-05 16:59:30'],
            [16, 2, 2, 2000, '2026-08-05 18:30:04'],
            [17, 2, 2, 2500, '2026-08-06 17:20:14'],
            [18, 2, 2, 4000, '2026-08-07 12:30:04'],
            [19, 2, 2, 1000, '2026-08-08 12:31:04'],
            [20, 2, 2, 7000, '2026-08-09 16:59:30'],
            [21, 2, 2, 1000, '2026-08-10 12:31:04'],
            [22, 2, 2, 7000, '2026-08-11 16:59:30'],
        ];

        foreach ($transactions as [$id, $merchantId, $outletId, $billTotal, $createdAt]) {
            $userId = $merchantId === 1 ? $user1->id : $user2->id;

            Transaction::updateOrCreate(
                ['id' => $id],
                [
                    'merchant_id' => $merchantId,
                    'outlet_id' => $outletId,
                    'bill_total' => $billTotal,
                    'created_at' => $createdAt,
                    'created_by' => $userId,
                    'updated_at' => $createdAt,
                    'updated_by' => $userId,
                ]
            );
        }
    }
}
