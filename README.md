# Sample go app
This is a sample go app that create a basic server along with prometheus metrics.

## CI/CD
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

## Trivy
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

## Additional configuration
### Docker
In order to use the advanced grafana dashboard filtering regarding the logs, you will need to add the following configuration to the docker daemon.

Edit the docker daemon (`/etc/docker/daemon.json`) to include the following:
(If the file does not exist, create it)
```json
{
  "log-opts": {
    "tag": "{{.ImageName}}|{{.Name}}|{{.ImageFullID}}|{{.FullID}}"
  }
}
```
Then restart the docker daemon:
```bash
sudo systemctl restart docker
```
