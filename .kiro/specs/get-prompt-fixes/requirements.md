# Requirements Document

## 1. Introduction

This document captures the product requirements for fixing the MCP `prompts/get` response so that prompts can be consumed without validation errors. The work is driven by epic `EP-011` and user story `US-053`, which report that the backend currently returns messages that violate the 18 June 2025 MCP Prompt Specification.

## 2. Problem Statement

When the MCP server responds to `prompts/get`, it emits message entries with `role: "system"` and serialises `content` as a plain string. The official specification requires message roles to be either `user` or `assistant`, and the `content` field to be an array of typed objects. Clients that follow the spec reject the response with validation errors (see requirement `REQ-057`) and fail to load the prompt payload.

## 3. Scope

- **In scope**
  - Bringing the `prompts/get` response schema in line with the MCP specification.
  - Introducing validation that prevents forbidden role values or malformed content structures from being emitted.
  - Extending the stored prompt message model to persist the role enum required by the specification.
  - Updating tests and documentation that cover prompt retrieval.
- **Out of scope**
  - Changes to other MCP tools (e.g. prompt updates, steering documents).
  - Altering the storage format of prompts inside the requirements system beyond what is needed for role persistence.

## 4. References

- MCP Server Prompt Specification (2025-06-18): `https://modelcontextprotocol.io/specification/2025-06-18/server/prompts.md`
- Epic `EP-011`: “Баг: MCP ошибка при получении промпта — некорректное значение enum и тип content”
- User story `US-053`: “Как MCP клиент, я хочу получать промпт без ошибок валидации”
- Requirement `REQ-057`: “Промпт сообщения возвращаются в допустимом формате MCP”
- Acceptance criteria `AC-014` 
```
GIVEN MCP сервер подготовлен для выдачи промптов,
WHEN клиент запрашивает prompts/get,
THEN ответ содержит массив messages,
AND каждый элемент messages имеет role со значением "assistant" или "user",
AND поле content каждого сообщения является объектом с полем "type" и допустимымым согласно спецификации значением "text" и полем "text" и значением из поля "description" модели Prompt,
AND в ответе отсутствуют элементы с role="system" или content в строковом формате
```

## 5. Glossary

- **MCP** – Model Context Protocol used to expose tools, resources, prompts, and instructions.
- **Prompt message** – An element of the `messages` array returned by `prompts/get`, consisting of a `role` and `content`.
- **Typed content chunk** – An object inside `message.content` that includes a `type` discriminator (e.g. `"text"`) and associated payload fields.

## 6. Functional Requirements

### FR1: Valid Roles in Prompt Messages

**Source:** `US-053`, MCP specification section “Prompt Message Roles”  
**Description:** Every message emitted in the `prompts/get` response SHALL use `role: "assistant"` or `role: "user"`. Any value outside the allowed enum (including `"system"`) SHALL be rejected before the response is returned to the client.  
**Acceptance:** GIVEN the MCP server is prepared to issue a prompt, WHEN `prompts/get` is invoked, THEN no message in the response contains an unauthorised role, satisfying `AC-014`.

### FR2: Structured Content Payload

**Source:** `REQ-057`, MCP specification section “Content Chunks”  
**Description:** Each message's `content` field SHALL be an array of objects. Every object SHALL include the `type` property (e.g. `"text"`) and additional fields mandated by the type definition. Raw strings SHALL NOT be emitted.  
**Acceptance:** WHEN `prompts/get` is invoked, THEN the response content validates against the 2025-06-18 schema and can be parsed by compliant clients without raising `invalid_type` errors.

### FR3: Guard Rails and Error Reporting

**Source:** Incident trace attached to `EP-011`  
**Description:** The server SHALL validate the prompt payload before serialisation. If violations are detected (e.g. forbidden role, malformed content chunk), the handler SHALL log the issue and return an MCP error (`Invalid params`) instead of propagating a malformed success response.  
**Acceptance:** WHEN the stored prompt contains invalid data, THEN `prompts/get` responds with a descriptive error and no longer emits the invalid payload.

### FR4: Persisted Prompt Message Role Field

**Source:** `EP-011`, MCP specification section “Prompt Message Roles”  
**Description:** The underlying prompt data model SHALL persist an explicit `role` attribute for every stored message so that downstream handlers can produce compliant `prompts/get` responses without inferring the role at runtime.  
**Acceptance:** WHEN prompt messages are retrieved from the database via existing services, THEN each message record exposes a `role` field holding a valid enum value (`assistant` or `user`).

## 7. Non-functional Requirements

- **Compatibility:** Changes MUST remain backward-compatible with existing prompt storage except where the stored data itself is invalid; the migration strategy will be covered in the design.
- **Observability:** Introduce structured logs or metrics to surface validation failures for operations and QA teams.
- **Testing:** Automated unit and integration tests SHALL cover both successful retrieval and failure scenarios described above.

## 8. Open Questions

1. Do we need to support additional content chunk types (e.g. `image`) beyond `"text"` for upcoming prompts?
2. Should invalid stored prompts trigger an automatic remediation workflow (e.g. data migration), or is failing the request sufficient?
