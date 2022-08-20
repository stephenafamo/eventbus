# EventBus

This is a package for peforming actions based on events. The flow is simple.

### 1. Create the payload

```go
type LoginPayload struct {
    ID    int64
    Email string
}
```

### 2. Create the event

Using an in-memory implemnetation

```go
memoryevents "github.com/stephenafamo/eventbus/memory"
loginEvent, _ :=  memoryevents.New[loginPayload](context.Background())
```

OR using a redis implemnetation based on [go-redis](https://github.com/go-redis/redis/v8)

```go
import redisevents "github.com/stephenafamo/eventbus/redis"
loginEvent, _ :=  redisevents.New[loginPayload](ctx, goRedisClient, "channelName")
```

### 3. Register Handlers

```go
func loginHandler1(loginPayload) {
    // Handle login event
}

// The ID is used if we want to UnregisterHandler the handler
loginEvent.RegisterHandler("someID", eventbus.HandlerFunc(loginHandler1))
```

### 4. Publish Events

```go
loginEvent.Publish(ctx, LoginPayload{
    ID:     10,
    Email: "stephen@example.com",
})
```

