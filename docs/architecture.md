# LangDuel Architecture

## High Level Overview

```mermaid
flowchart LR
    Browser -->|WebSocket| Server
    Server --> Hub
    Hub --> Clients
```sssssssssssssssssssssss