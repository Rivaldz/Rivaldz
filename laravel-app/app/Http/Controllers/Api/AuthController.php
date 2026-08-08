<?php

namespace App\Http\Controllers\Api;

use App\Http\Controllers\Controller;
use App\Models\Merchant;
use App\Models\User;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Auth;
use Illuminate\Support\Facades\DB;
use OpenApi\Attributes as OA;
use Tymon\JWTAuth\JWTGuard;

class AuthController extends Controller
{
    /**
     * POST /api/auth/register
     *
     * Register a new user along with their merchant account.
     */
    #[OA\Post(
        path: '/auth/register',
        operationId: 'authRegister',
        summary: 'Register a new user and merchant',
        description: 'Create a new user account and its associated merchant. After registration, login via POST /api/auth/login to get a JWT token.',
        tags: ['Authentication'],
        requestBody: new OA\RequestBody(
            required: true,
            content: new OA\JsonContent(
                required: ['name', 'email', 'password', 'password_confirmation', 'merchant_name'],
                properties: [
                    new OA\Property(property: 'name', type: 'string', maxLength: 255, example: 'John Doe'),
                    new OA\Property(property: 'email', type: 'string', format: 'email', example: 'john@example.com'),
                    new OA\Property(property: 'password', type: 'string', format: 'password', minLength: 8, example: 'secret123'),
                    new OA\Property(property: 'password_confirmation', type: 'string', format: 'password', example: 'secret123'),
                    new OA\Property(property: 'merchant_name', type: 'string', maxLength: 40, example: "John's Store"),
                ]
            )
        ),
        responses: [
            new OA\Response(
                response: 201,
                description: 'Registration successful',
                content: new OA\JsonContent(
                    properties: [
                        new OA\Property(property: 'success', type: 'boolean', example: true),
                        new OA\Property(property: 'message', type: 'string', example: 'Registration successful. Please login to get your access token.'),
                        new OA\Property(
                            property: 'data',
                            type: 'object',
                            properties: [
                                new OA\Property(property: 'id', type: 'integer', example: 3),
                                new OA\Property(property: 'name', type: 'string', example: 'John Doe'),
                                new OA\Property(property: 'email', type: 'string', format: 'email', example: 'john@example.com'),
                                new OA\Property(
                                    property: 'merchant',
                                    type: 'object',
                                    properties: [
                                        new OA\Property(property: 'id', type: 'integer', example: 3),
                                        new OA\Property(property: 'merchant_name', type: 'string', example: "John's Store"),
                                    ]
                                ),
                            ]
                        ),
                    ]
                )
            ),
            new OA\Response(
                response: 422,
                description: 'Validation error',
                content: new OA\JsonContent(
                    properties: [
                        new OA\Property(property: 'message', type: 'string', example: 'The email has already been taken.'),
                        new OA\Property(property: 'errors', type: 'object'),
                    ]
                )
            ),
        ]
    )]
    public function register(Request $request): JsonResponse
    {
        $validated = $request->validate([
            'name'          => ['required', 'string', 'max:255'],
            'email'         => ['required', 'string', 'email', 'unique:users,email'],
            'password'      => ['required', 'string', 'min:8', 'confirmed'],
            'merchant_name' => ['required', 'string', 'max:40'],
        ]);

        [$user, $merchant] = DB::transaction(function () use ($validated) {
            $user = User::create([
                'name'     => $validated['name'],
                'email'    => $validated['email'],
                'password' => $validated['password'],
            ]);

            $merchant = Merchant::create([
                'user_id'       => $user->id,
                'merchant_name' => $validated['merchant_name'],
                'created_by'    => $user->id,
                'updated_by'    => $user->id,
            ]);

            return [$user, $merchant];
        });

        return response()->json([
            'success' => true,
            'message' => 'Registration successful. Please login to get your access token.',
            'data'    => [
                'id'       => $user->id,
                'name'     => $user->name,
                'email'    => $user->email,
                'merchant' => [
                    'id'            => $merchant->id,
                    'merchant_name' => $merchant->merchant_name,
                ],
            ],
        ], 201);
    }

    /**
     * POST /api/auth/login
     *
     * Authenticate a user and return a JWT access token.
     */
    #[OA\Post(
        path: '/auth/login',
        operationId: 'authLogin',
        summary: 'Login and get JWT token',
        description: 'Authenticate a user with email and password, returns a JWT access token.',
        tags: ['Authentication'],
        requestBody: new OA\RequestBody(
            required: true,
            content: new OA\JsonContent(
                required: ['email', 'password'],
                properties: [
                    new OA\Property(property: 'email', type: 'string', format: 'email', example: 'merchant1@example.com'),
                    new OA\Property(property: 'password', type: 'string', format: 'password', example: 'password'),
                ]
            )
        ),
        responses: [
            new OA\Response(
                response: 200,
                description: 'Login successful',
                content: new OA\JsonContent(
                    properties: [
                        new OA\Property(property: 'success', type: 'boolean', example: true),
                        new OA\Property(property: 'token', type: 'string', example: 'eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9...'),
                        new OA\Property(property: 'token_type', type: 'string', example: 'bearer'),
                        new OA\Property(property: 'expires_in', type: 'integer', example: 3600),
                    ]
                )
            ),
            new OA\Response(
                response: 401,
                description: 'Invalid credentials',
                content: new OA\JsonContent(
                    properties: [
                        new OA\Property(property: 'success', type: 'boolean', example: false),
                        new OA\Property(property: 'message', type: 'string', example: 'Invalid credentials.'),
                    ]
                )
            ),
            new OA\Response(
                response: 422,
                description: 'Validation error',
                content: new OA\JsonContent(
                    properties: [
                        new OA\Property(property: 'message', type: 'string', example: 'The email field is required.'),
                        new OA\Property(property: 'errors', type: 'object'),
                    ]
                )
            ),
        ]
    )]
    public function login(Request $request): JsonResponse
    {
        $credentials = $request->validate([
            'email' => ['required', 'email'],
            'password' => ['required', 'string'],
        ]);

        if (! $token = Auth::guard('api')->attempt($credentials)) {
            return response()->json([
                'success' => false,
                'message' => 'Invalid credentials.',
            ], 401);
        }

        return $this->respondWithToken($token);
    }

    /**
     * POST /api/auth/logout
     *
     * Invalidate the current JWT.
     */
    #[OA\Post(
        path: '/auth/logout',
        operationId: 'authLogout',
        summary: 'Logout (invalidate JWT)',
        description: 'Invalidate the current JWT token.',
        security: [['bearerAuth' => []]],
        tags: ['Authentication'],
        responses: [
            new OA\Response(
                response: 200,
                description: 'Successfully logged out',
                content: new OA\JsonContent(
                    properties: [
                        new OA\Property(property: 'success', type: 'boolean', example: true),
                        new OA\Property(property: 'message', type: 'string', example: 'Successfully logged out.'),
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
        ]
    )]
    public function logout(): JsonResponse
    {
        Auth::guard('api')->logout();

        return response()->json([
            'success' => true,
            'message' => 'Successfully logged out.',
        ], 200);
    }

    /**
     * POST /api/auth/refresh
     *
     * Exchange an expired but valid token for a new one.
     */
    #[OA\Post(
        path: '/auth/refresh',
        operationId: 'authRefresh',
        summary: 'Refresh JWT token',
        description: 'Exchange an expired but valid JWT token for a new one.',
        security: [['bearerAuth' => []]],
        tags: ['Authentication'],
        responses: [
            new OA\Response(
                response: 200,
                description: 'Token refreshed successfully',
                content: new OA\JsonContent(
                    properties: [
                        new OA\Property(property: 'success', type: 'boolean', example: true),
                        new OA\Property(property: 'token', type: 'string', example: 'eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9...'),
                        new OA\Property(property: 'token_type', type: 'string', example: 'bearer'),
                        new OA\Property(property: 'expires_in', type: 'integer', example: 3600),
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
        ]
    )]
    public function refresh(): JsonResponse
    {
        /** @var JWTGuard $guard */
        $guard = Auth::guard('api');

        return $this->respondWithToken($guard->refresh());
    }

    /**
     * GET /api/auth/me
     *
     * Return the authenticated user's profile (plus their merchant context).
     */
    #[OA\Get(
        path: '/auth/me',
        operationId: 'authMe',
        summary: 'Get current user profile',
        description: "Return the authenticated user's profile including their merchant context.",
        security: [['bearerAuth' => []]],
        tags: ['Authentication'],
        responses: [
            new OA\Response(
                response: 200,
                description: 'User profile retrieved',
                content: new OA\JsonContent(
                    properties: [
                        new OA\Property(property: 'success', type: 'boolean', example: true),
                        new OA\Property(
                            property: 'data',
                            type: 'object',
                            properties: [
                                new OA\Property(property: 'id', type: 'integer', example: 1),
                                new OA\Property(property: 'name', type: 'string', example: 'Merchant 1 Owner'),
                                new OA\Property(property: 'email', type: 'string', format: 'email', example: 'merchant1@example.com'),
                                new OA\Property(
                                    property: 'merchant',
                                    type: 'object',
                                    nullable: true,
                                    properties: [
                                        new OA\Property(property: 'id', type: 'integer', example: 1),
                                        new OA\Property(property: 'merchant_name', type: 'string', example: 'Merchant 1'),
                                    ]
                                ),
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
        ]
    )]
    public function me(Request $request): JsonResponse
    {
        /** @var \App\Models\User $user */
        $user = $request->user('api');

        return response()->json([
            'success' => true,
            'data' => [
                'id' => $user->id,
                'name' => $user->name,
                'email' => $user->email,
                'merchant' => $user->merchant
                    ? [
                        'id' => $user->merchant->id,
                        'merchant_name' => $user->merchant->merchant_name,
                    ]
                    : null,
            ],
        ], 200);
    }

    protected function respondWithToken(string $token): JsonResponse
    {
        /** @var JWTGuard $guard */
        $guard = Auth::guard('api');

        return response()->json([
            'success' => true,
            'token' => $token,
            'token_type' => 'bearer',
            'expires_in' => $guard->factory()->getTTL() * 60,
        ], 200);
    }
}

