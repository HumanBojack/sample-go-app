# Sample go app
This is a sample go app that create a basic server along with prometheus metrics.

### CI/CD
Here is a diagram of the CI/CD pipeline

```mermaid
sequenceDiagram
    Dev->>+Github: commit
    Github->>+Github: Check for vulnerabilities
    Github->>Github: If vulnerability scan fails
    alt Vulnerability scan fails
        Note over Github: Stop workflow
    else Vulnerability scan succeeds
        Github->>+Github: Build image
        Github->>+Ghcr: push image
    end
```