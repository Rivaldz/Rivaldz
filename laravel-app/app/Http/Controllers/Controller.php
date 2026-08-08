<?php

namespace App\Http\Controllers;

use OpenApi\Attributes as OA;

#[OA\Info(
    version: '1.0.0',
    title: 'Majoo Revenue Reporting API',
    description: 'Laravel API for Majoo PHP programming test — JWT authentication, daily revenue (omzet) reporting per merchant and outlet, with multi-tenancy.',
    contact: new OA\Contact(email: 'admin@majoo.id')
)]
#[OA\Server(url: '/api', description: 'API Server')]
#[OA\SecurityScheme(
    securityScheme: 'bearerAuth',
    type: 'http',
    scheme: 'bearer',
    bearerFormat: 'JWT',
    description: 'Enter JWT Bearer token. Login via POST /api/auth/login to get the token.'
)]
#[OA\Tag(name: 'Authentication', description: 'JWT authentication endpoints')]
#[OA\Tag(name: 'Reports', description: 'Revenue reporting endpoints')]
abstract class Controller
{
    //
}


