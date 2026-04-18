---
name: ROUTER_MODULE_GUIDE
description: Router Module Configuration Guide
---
# Router Module Configuration Guide

## Problem

When using `RouterModule.register()` for path prefixing, controllers inside nested imported modules are **NOT recognized**. 

### Example of the Issue:

```typescript
// ❌ WRONG - Controller won't be recognized by RouterModule
@Module({
    imports: [
        LibUploadS3Module.registerAsync({...}),  // Has UploadS3Controller
    ],
    controllers: [],  // Empty - Router won't find controller
})
export class UploadPortalPresentationModule {}

// Router configuration
{
    path: 'upload',
    module: UploadPortalPresentationModule,  // Expected: /v1/portal/upload
}
// Result: Controller routes not prefixed correctly ❌
```

## Solution

**Controllers must be declared directly in the module** that's registered with RouterModule.

### Fixed Implementation:

#### 1. Remove Controller from Library Module

```typescript
// libs/src/core/upload/s3/upload-s3.module.ts
@Module({})
export class LibUploadS3Module {
    public static registerAsync(configure: UploadS3AsyncConfigure): DynamicModule {
        return {
            module: LibUploadS3Module,
            providers: [...],
            controllers: [], // ✅ Empty - don't register controllers here
            exports: [...],
        };
    }
}
```

#### 2. Register Controller in Presentation Module

```typescript
// src/presentation/portal/upload/upload.module.ts
import { UploadS3Controller } from '@core/upload/s3';

@Module({
    imports: [],
    controllers: [UploadS3Controller], // ✅ Register controller here
    providers: [],
})
export class UploadPortalPresentationModule {}
```

#### 3. Router Configuration

```typescript
// src/presentation/portal/portal.module.ts
const routers: Routes = [
    {
        path: 'portal',
        children: [
            {
                path: 'upload',
                module: UploadPortalPresentationModule,
            },
            // ... other routes
        ],
    },
];

@Module({
    imports: [
        LibRouterModule.register(routers),
        UploadS3InfrastructureModule, // Provides services/providers
        // ... other modules
    ],
})
export class PortalModule {}
```

## Result

✅ Routes are now correctly prefixed:
- `POST /v1/portal/upload` - Upload files endpoint
- All controllers in `UploadPortalPresentationModule` get the correct path prefix

## Architecture Explanation

### Why This Works

```
┌─────────────────────────────────────────────────────────────┐
│                    RouterModule                              │
│  - Scans modules in routes array                            │
│  - Only recognizes controllers in those modules             │
│  - Does NOT scan nested imports                             │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│         UploadPortalPresentationModule                       │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  controllers: [UploadS3Controller] ✅                 │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### Separation of Concerns

1. **Library Module (`LibUploadS3Module`)**
   - Provides services, providers, adapters
   - Exports reusable business logic
   - **No controllers** - presentation layer independent

2. **Infrastructure Module (`UploadS3InfrastructureModule`)**
   - Configures library with environment-specific settings
   - Registers multiple scopes (PUBLIC, PRIVATE)
   - **No controllers** - just configuration

3. **Presentation Module (`UploadPortalPresentationModule`)**
   - **Registers controllers** for specific route prefix
   - Can have its own DTOs, validation
   - Presentation layer responsibility

## Best Practices

### ✅ DO

```typescript
// Presentation module owns its controllers
@Module({
    imports: [SomeLibraryModule],
    controllers: [MyController], // ✅
})
export class MyPresentationModule {}
```

### ❌ DON'T

```typescript
// Library module should not have controllers
export class LibraryModule {
    static forRoot() {
        return {
            controllers: [LibraryController], // ❌
        };
    }
}
```

## Multiple Presentation Modules

You can use the same library in different presentation contexts:

```typescript
// Portal presentation
@Module({
    controllers: [UploadS3Controller],
})
export class UploadPortalPresentationModule {}

// Public API presentation
@Module({
    controllers: [UploadS3Controller],
})
export class UploadPublicPresentationModule {}

// Internal API presentation
@Module({
    controllers: [UploadS3Controller],
})
export class UploadInternalPresentationModule {}

// Router configuration
const routers: Routes = [
    { path: 'portal/upload', module: UploadPortalPresentationModule },
    { path: 'public/upload', module: UploadPublicPresentationModule },
    { path: 'internal/upload', module: UploadInternalPresentationModule },
];
```

Result:
- `POST /v1/portal/upload` - Portal upload
- `POST /v1/public/upload` - Public upload  
- `POST /v1/internal/upload` - Internal upload

## Debugging Tips

### Check registered routes

```bash
# Enable NestJS debug logging
LOG_LEVEL=debug npm run start:dev
```

Look for logs like:
```
[RouterExplorer] Mapped {/v1/portal/upload, POST} route
```

### Verify controller registration

```typescript
import { NestFactory } from '@nestjs/core';

async function bootstrap() {
    const app = await NestFactory.create(AppModule);
    
    // Log all routes
    const server = app.getHttpServer();
    const router = server._events.request._router;
    
    console.log(router.stack
        .filter(r => r.route)
        .map(r => ({
            path: r.route.path,
            methods: Object.keys(r.route.methods),
        }))
    );
}
```

## Common Mistakes

### 1. Controller in wrong module
```typescript
// ❌ Controller in library module
LibUploadS3Module.registerAsync({
    controllers: [UploadS3Controller],  // Wrong!
});

// ✅ Controller in presentation module
@Module({
    controllers: [UploadS3Controller],  // Correct!
})
export class UploadPortalPresentationModule {}
```

### 2. Forgetting to import library
```typescript
// ❌ No library imports
@Module({
    controllers: [UploadS3Controller],  // Controller exists
    // But no providers/services available!
})

// ✅ Import library for services
@Module({
    imports: [UploadS3InfrastructureModule],  // Provides services
    controllers: [UploadS3Controller],
})
```

### 3. Wrong router configuration
```typescript
// ❌ Wrong - using infrastructure module
{
    path: 'upload',
    module: UploadS3InfrastructureModule,  // Has no controllers!
}

// ✅ Correct - using presentation module
{
    path: 'upload',
    module: UploadPortalPresentationModule,  // Has controllers!
}
```

## Related Files

- `/libs/src/core/upload/s3/upload-s3.module.ts` - Library module (no controllers)
- `/libs/src/core/upload/s3/upload-s3.controller.ts` - Shared controller
- `/src/infrastructure/upload-s3-service/upload-s3.module.ts` - Infrastructure setup
- `/src/presentation/portal/upload/upload.module.ts` - Presentation module (has controllers)
- `/src/presentation/portal/portal.module.ts` - Router configuration





