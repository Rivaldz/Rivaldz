<?php

namespace App\Http\Resources;

use Illuminate\Http\Request;
use Illuminate\Http\Resources\Json\JsonResource;

/**
 * Represents a single daily revenue entry.
 *
 * Wrapping a LengthAwarePaginator in a resource collection produces the
 * standard Laravel paginated shape: { data: [...], links: {...}, meta: {...} }.
 */
class DailyRevenueResource extends JsonResource
{
    public function toArray(Request $request): array
    {
        return [
            'date' => $this['date'],
            'revenue' => round($this['revenue'], 2),
        ];
    }
}
