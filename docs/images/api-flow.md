sequenceDiagram
    participant C as Client
    participant A as API Gateway
    participant S as Service
    participant R as Redis
    participant D as Database
    
    Note over C,D: CREATE SHORT URL FLOW
    C->>A: POST /api/v1/shorten
    A->>S: Validate & Process
    S->>D: Store URL
    D-->>S: Confirm
    S->>R: Cache Response
    S-->>A: Success Response
    A-->>C: Short URL Created
    
    Note over C,D: REDIRECT FLOW
    C->>A: GET /:code
    A->>S: Process Request
    S->>R: Check Cache
    alt Cache Hit
        R-->>S: Cached Data
        S->>D: Record Click (Async)
    else Cache Miss
        S->>D: Fetch URL
        D-->>S: URL Data
        S->>R: Update Cache
        S->>D: Record Click
    end
    S-->>A: Redirect Response
    A-->>C: 301 Redirect

**Key Features:**
- Cache-first strategy for performance
- Asynchronous click tracking
- Automatic cache warming
- Graceful fallback to database
