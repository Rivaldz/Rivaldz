<?php

namespace App\Http\Controllers\Api;

use App\Http\Controllers\Controller;
use App\Http\Resources\DailyRevenueResource;
use App\Models\Merchant;
use App\Models\Outlet;
use App\Models\Transaction;
use Illuminate\Http\Request;
use Illuminate\Pagination\LengthAwarePaginator;
use Illuminate\Support\Carbon;
use Illuminate\Support\Collection;
use OpenApi\Attributes as OA;

class ReportController extends Controller
{
    /**
     * GET /api/reports/merchant/daily
     *
     * Daily revenue (omzet = SUM of bill_total) for the authenticated user's
     * merchant, grouped per day for the requested month/year. Days without any
     * transaction are reported with a revenue of 0.
     *
     * Default month/year is November 2026 as per the test spec.
     */
    #[OA\Get(
        path: '/reports/merchant/daily',
        operationId: 'reportMerchantDaily',
        summary: 'Merchant daily revenue report',
        description: "Daily revenue (omzet) for the authenticated user's merchant. Days without transactions show revenue = 0. Default: November 2026.",
        security: [['bearerAuth' => []]],
        tags: ['Reports'],
        parameters: [
            new OA\Parameter(
                name: 'month',
                in: 'query',
                required: false,
                description: 'Report month (1-12). Default: 11',
                schema: new OA\Schema(type: 'integer', minimum: 1, maximum: 12, example: 11)
            ),
            new OA\Parameter(
                name: 'year',
                in: 'query',
                required: false,
                description: 'Report year (2000-2100). Default: 2026',
                schema: new OA\Schema(type: 'integer', minimum: 2000, maximum: 2100, example: 2026)
            ),
            new OA\Parameter(
                name: 'page',
                in: 'query',
                required: false,
                description: 'Pagination page number. Default: 1',
                schema: new OA\Schema(type: 'integer', minimum: 1, example: 1)
            ),
            new OA\Parameter(
                name: 'per_page',
                in: 'query',
                required: false,
                description: 'Items per page (1-100). Default: 10',
                schema: new OA\Schema(type: 'integer', minimum: 1, maximum: 100, example: 10)
            ),
        ],
        responses: [
            new OA\Response(
                response: 200,
                description: 'Daily revenue report (paginated)',
                content: new OA\JsonContent(
                    properties: [
                        new OA\Property(
                            property: 'data',
                            type: 'array',
                            items: new OA\Items(
                                properties: [
                                    new OA\Property(property: 'date', type: 'string', format: 'date', example: '2026-11-01'),
                                    new OA\Property(property: 'revenue', type: 'number', format: 'float', example: 150000.00),
                                ]
                            )
                        ),
                        new OA\Property(
                            property: 'links',
                            type: 'object',
                            properties: [
                                new OA\Property(property: 'first', type: 'string', example: 'http://localhost:8000/api/reports/merchant/daily?page=1'),
                                new OA\Property(property: 'last', type: 'string', example: 'http://localhost:8000/api/reports/merchant/daily?page=3'),
                                new OA\Property(property: 'prev', type: 'string', nullable: true, example: null),
                                new OA\Property(property: 'next', type: 'string', nullable: true, example: 'http://localhost:8000/api/reports/merchant/daily?page=2'),
                            ]
                        ),
                        new OA\Property(
                            property: 'meta',
                            type: 'object',
                            properties: [
                                new OA\Property(property: 'current_page', type: 'integer', example: 1),
                                new OA\Property(property: 'from', type: 'integer', example: 1),
                                new OA\Property(property: 'last_page', type: 'integer', example: 3),
                                new OA\Property(property: 'per_page', type: 'integer', example: 10),
                                new OA\Property(property: 'to', type: 'integer', example: 10),
                                new OA\Property(property: 'total', type: 'integer', example: 30),
                            ]
                        ),
                    ]
                )
            ),
            new OA\Response(
                response: 401,
                description: 'Unauthenticated',
                content: new OA\JsonContent(
                    properties: [
                        new OA\Property(property: 'message', type: 'string', example: 'Unauthenticated.'),
                    ]
                )
            ),
            new OA\Response(
                response: 422,
                description: 'Validation error',
                content: new OA\JsonContent(
                    properties: [
                        new OA\Property(property: 'message', type: 'string', example: 'The month field must be between 1 and 12.'),
                        new OA\Property(property: 'errors', type: 'object'),
                    ]
                )
            ),
        ]
    )]
    public function merchantDaily(Request $request)
    {
        $validated = $request->validate([
            'month' => ['nullable', 'integer', 'between:1,12'],
            'year' => ['nullable', 'integer', 'between:2000,2100'],
            'per_page' => ['nullable', 'integer', 'between:1,100'],
        ]);

        /** @var Merchant $merchant */
        $merchant = $request->attributes->get('tenant_merchant');

        $month = $validated['month'] ?? 11;
        $year = $validated['year'] ?? 2026;

        $report = $this->dailyRevenue(
            scope: fn ($query) => $query->where('merchant_id', $merchant->id),
            month: $month,
            year: $year,
        );

        return DailyRevenueResource::collection(
            $this->paginate($request, $report)
        );
    }

    /**
     * GET /api/reports/outlet/daily
     *
     * Daily revenue for a single outlet owned by the authenticated merchant.
     * Returns 403 when the outlet belongs to another merchant (multi-tenancy).
     *
     * Default month/year is August 2026 as per the test spec.
     */
    #[OA\Get(
        path: '/reports/outlet/daily',
        operationId: 'reportOutletDaily',
        summary: 'Outlet daily revenue report',
        description: 'Daily revenue for a specific outlet owned by the authenticated merchant. Returns 403 if the outlet does not belong to your merchant. Default: August 2026.',
        security: [['bearerAuth' => []]],
        tags: ['Reports'],
        parameters: [
            new OA\Parameter(
                name: 'outlet_id',
                in: 'query',
                required: true,
                description: 'ID of the outlet (must belong to your merchant)',
                schema: new OA\Schema(type: 'integer', example: 1)
            ),
            new OA\Parameter(
                name: 'month',
                in: 'query',
                required: false,
                description: 'Report month (1-12). Default: 8',
                schema: new OA\Schema(type: 'integer', minimum: 1, maximum: 12, example: 8)
            ),
            new OA\Parameter(
                name: 'year',
                in: 'query',
                required: false,
                description: 'Report year (2000-2100). Default: 2026',
                schema: new OA\Schema(type: 'integer', minimum: 2000, maximum: 2100, example: 2026)
            ),
            new OA\Parameter(
                name: 'page',
                in: 'query',
                required: false,
                description: 'Pagination page number. Default: 1',
                schema: new OA\Schema(type: 'integer', minimum: 1, example: 1)
            ),
            new OA\Parameter(
                name: 'per_page',
                in: 'query',
                required: false,
                description: 'Items per page (1-100). Default: 10',
                schema: new OA\Schema(type: 'integer', minimum: 1, maximum: 100, example: 10)
            ),
        ],
        responses: [
            new OA\Response(
                response: 200,
                description: 'Daily revenue report (paginated)',
                content: new OA\JsonContent(
                    properties: [
                        new OA\Property(
                            property: 'data',
                            type: 'array',
                            items: new OA\Items(
                                properties: [
                                    new OA\Property(property: 'date', type: 'string', format: 'date', example: '2026-08-01'),
                                    new OA\Property(property: 'revenue', type: 'number', format: 'float', example: 75000.50),
                                ]
                            )
                        ),
                        new OA\Property(
                            property: 'links',
                            type: 'object',
                            properties: [
                                new OA\Property(property: 'first', type: 'string', example: 'http://localhost:8000/api/reports/outlet/daily?page=1'),
                                new OA\Property(property: 'last', type: 'string', example: 'http://localhost:8000/api/reports/outlet/daily?page=4'),
                                new OA\Property(property: 'prev', type: 'string', nullable: true, example: null),
                                new OA\Property(property: 'next', type: 'string', nullable: true, example: 'http://localhost:8000/api/reports/outlet/daily?page=2'),
                            ]
                        ),
                        new OA\Property(
                            property: 'meta',
                            type: 'object',
                            properties: [
                                new OA\Property(property: 'current_page', type: 'integer', example: 1),
                                new OA\Property(property: 'from', type: 'integer', example: 1),
                                new OA\Property(property: 'last_page', type: 'integer', example: 4),
                                new OA\Property(property: 'per_page', type: 'integer', example: 10),
                                new OA\Property(property: 'to', type: 'integer', example: 10),
                                new OA\Property(property: 'total', type: 'integer', example: 31),
                            ]
                        ),
                    ]
                )
            ),
            new OA\Response(
                response: 401,
                description: 'Unauthenticated',
                content: new OA\JsonContent(
                    properties: [
                        new OA\Property(property: 'message', type: 'string', example: 'Unauthenticated.'),
                    ]
                )
            ),
            new OA\Response(
                response: 403,
                description: 'Outlet does not belong to your merchant',
                content: new OA\JsonContent(
                    properties: [
                        new OA\Property(property: 'success', type: 'boolean', example: false),
                        new OA\Property(property: 'message', type: 'string', example: 'The outlet does not belong to your merchant account.'),
                    ]
                )
            ),
            new OA\Response(
                response: 422,
                description: 'Validation error',
                content: new OA\JsonContent(
                    properties: [
                        new OA\Property(property: 'message', type: 'string', example: 'The outlet id field is required.'),
                        new OA\Property(property: 'errors', type: 'object'),
                    ]
                )
            ),
        ]
    )]
    public function outletDaily(Request $request)
    {
        $validated = $request->validate([
            'outlet_id' => ['required', 'integer'],
            'month' => ['nullable', 'integer', 'between:1,12'],
            'year' => ['nullable', 'integer', 'between:2000,2100'],
            'per_page' => ['nullable', 'integer', 'between:1,100'],
        ]);

        /** @var Merchant $merchant */
        $merchant = $request->attributes->get('tenant_merchant');

        $outlet = Outlet::query()
            ->withoutGlobalScopes()
            ->where('id', $validated['outlet_id'])
            ->first();

        if (! $outlet || $outlet->merchant_id !== $merchant->id) {
            return response()->json([
                'success' => false,
                'message' => 'The outlet does not belong to your merchant account.',
            ], 403);
        }

        $month = $validated['month'] ?? 8;
        $year = $validated['year'] ?? 2026;

        $report = $this->dailyRevenue(
            scope: fn ($query) => $query->where('outlet_id', $outlet->id),
            month: $month,
            year: $year,
        );

        return DailyRevenueResource::collection(
            $this->paginate($request, $report)
        );
    }

    /**
     * Build a day-by-day revenue map for the given month/year.
     *
     * Generates the full calendar for the month, then left-fills revenue from
     * the aggregated transactions so dates without transactions show "0".
     */
    protected function dailyRevenue(callable $scope, int $month, int $year): Collection
    {
        $start = Carbon::create($year, $month, 1)->startOfDay();
        $end = $start->copy()->endOfMonth()->endOfDay();

        $days = collect();
        for ($day = $start->copy(); $day->lte($end); $day->addDay()) {
            $days[$day->format('Y-m-d')] = 0.0;
        }

        $rows = Transaction::query()
            ->selectRaw('DATE(created_at) as date')
            ->selectRaw('SUM(bill_total) as revenue')
            ->whereBetween('created_at', [$start, $end])
            ->tap($scope)
            ->groupByRaw('DATE(created_at)')
            ->get();

        foreach ($rows as $row) {
            if ($days->has($row->date)) {
                $days[$row->date] = (float) $row->revenue;
            }
        }

        return $days
            ->sortKeys()
            ->map(fn (float $revenue, string $date) => [
                'date' => $date,
                'revenue' => $revenue,
            ])
            ->values();
    }

    protected function paginate(Request $request, Collection $items): LengthAwarePaginator
    {
        $perPage = (int) $request->input('per_page', 10);
        $page = LengthAwarePaginator::resolveCurrentPage();

        return new LengthAwarePaginator(
            $items->forPage($page, $perPage),
            $items->count(),
            $perPage,
            $page,
            [
                'path' => $request->url(),
                'query' => $request->query(),
            ]
        );
    }
}

