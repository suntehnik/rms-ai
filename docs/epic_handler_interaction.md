### Сценарий 1: Получение Epic по UUID

```mermaid
sequenceDiagram
    actor Client
    participant Router as Gin Router
    participant Handler as EpicHandler
    participant Service as EpicService
    participant Repository as EpicRepository

    Client->>+Router: GET /api/v1/epics/{uuid}
    Router->>+Handler: GetEpic(context)
    Handler->>Handler: id = context.Param("id")
    Handler->>+Service: GetEpicByID(id)
    Service->>+Repository: FindByID(id)
    Repository-->>-Service: returns epic data
    Service-->>-Handler: returns epic model
    
    alt Epic Found
        Handler->>-Router: JSON(200, epic model)
    else Epic Not Found
        Handler->>-Router: JSON(404, "Not Found" error)
    end
    
    Router-->>-Client: HTTP Response
```

### Сценарий 2: Получение Epic по Reference ID

```mermaid
sequenceDiagram
    actor Client
    participant Router as Gin Router
    participant Handler as EpicHandler
    participant Service as EpicService
    participant Repository as EpicRepository

    Client->>+Router: GET /api/v1/epics/{reference-id}
    Router->>+Handler: GetEpic(context)
    Handler->>Handler: id = context.Param("id")
    Handler->>+Service: GetEpicByReferenceID(id)
    Service->>+Repository: FindByReferenceID(id)
    Repository-->>-Service: returns epic data
    Service-->>-Handler: returns epic model

    alt Epic Found
        Handler->>-Router: JSON(200, epic model)
    else Epic Not Found
        Handler->>-Router: JSON(404, "Not Found" error)
    end

    Router-->>-Client: HTTP Response
```
