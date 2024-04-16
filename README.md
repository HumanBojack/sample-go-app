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

### Trivy
We run a daily check for vulnerabilities using trivy.
```mermaid
sequenceDiagram
    cron->>+Github: Every week day
    Github->>+Github: Checkout code (main branch)
    Github->>+Github: Run Trivy vulnerability scanner
    alt Vulnerabilities found
        Note over Github: Exit with code 1
    else No vulnerabilities found
        Note over Github: Exit with code 0
    end
```