# orzkratos
Make it simple to execute Kratos commands.

With the following features:
1. Make it simple to execute `kratos proto add api/helloworld/demo.proto`
2. Sync Service with Proto: When the proto file is changed, the service will auto update with new methods.

# Install

```bash
go install github.com/orzkratos/orzkratos/cmd/orzkratos-add-proto@latest
go install github.com/orzkratos/orzkratos/cmd/orzkratos-srv-proto@latest
```

# Commands

## Add Proto

### New demo proto:
```bash
cd project-path/api/helloworld && orzkratos-add-proto demo.proto
```

Same as:
```bash
cd project-path && kratos proto add api/helloworld/demo.proto
```

### Simple command:
```bash
cd project-path/api/helloworld

orzkratos-add-proto demo
```

## Sync Service with Proto

### Sync demo service with proto:
```bash
cd project-path/api/helloworld && orzkratos-srv-proto demo.proto
```

Same as:
```bash
cd project-path && orzkratos-srv-proto api/helloworld/demo.proto
```

### Simple command:
```bash
cd project-path/api/helloworld

orzkratos-srv-proto
```
