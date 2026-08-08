<?php

namespace Tests\Feature;

use App\Models\User;
use Illuminate\Foundation\Testing\RefreshDatabase;
use Tests\TestCase;

class RegisterTest extends TestCase
{
    use RefreshDatabase;

    public function test_register_creates_user_and_merchant(): void
    {
        $response = $this->postJson('/api/auth/register', [
            'name'                  => 'Test User',
            'email'                 => 'test@example.com',
            'password'              => 'password123',
            'password_confirmation' => 'password123',
            'merchant_name'         => 'Test Merchant',
        ]);

        $response->assertStatus(201)
            ->assertJsonStructure([
                'success',
                'message',
                'data' => ['id', 'name', 'email', 'merchant' => ['id', 'merchant_name']],
            ])
            ->assertJson([
                'success' => true,
                'data' => [
                    'name'     => 'Test User',
                    'email'    => 'test@example.com',
                    'merchant' => ['merchant_name' => 'Test Merchant'],
                ],
            ]);

        $this->assertDatabaseHas('users', ['email' => 'test@example.com']);
        $this->assertDatabaseHas('merchants', ['merchant_name' => 'Test Merchant']);
    }

    public function test_register_fails_with_duplicate_email(): void
    {
        User::factory()->create(['email' => 'taken@example.com']);

        $response = $this->postJson('/api/auth/register', [
            'name'                  => 'Another User',
            'email'                 => 'taken@example.com',
            'password'              => 'password123',
            'password_confirmation' => 'password123',
            'merchant_name'         => 'Some Merchant',
        ]);

        $response->assertStatus(422)
            ->assertJsonValidationErrors(['email']);
    }

    public function test_register_fails_without_required_fields(): void
    {
        $response = $this->postJson('/api/auth/register', []);

        $response->assertStatus(422)
            ->assertJsonValidationErrors(['name', 'email', 'password', 'merchant_name']);
    }

    public function test_register_fails_with_password_mismatch(): void
    {
        $response = $this->postJson('/api/auth/register', [
            'name'                  => 'Test User',
            'email'                 => 'test@example.com',
            'password'              => 'password123',
            'password_confirmation' => 'different_password',
            'merchant_name'         => 'Test Merchant',
        ]);

        $response->assertStatus(422)
            ->assertJsonValidationErrors(['password']);
    }

    public function test_registered_user_can_login(): void
    {
        $this->postJson('/api/auth/register', [
            'name'                  => 'Test User',
            'email'                 => 'test@example.com',
            'password'              => 'password123',
            'password_confirmation' => 'password123',
            'merchant_name'         => 'Test Merchant',
        ])->assertStatus(201);

        $loginResponse = $this->postJson('/api/auth/login', [
            'email'    => 'test@example.com',
            'password' => 'password123',
        ]);

        $loginResponse->assertStatus(200)
            ->assertJsonStructure(['success', 'token', 'token_type', 'expires_in']);
    }
}
