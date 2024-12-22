# clikratos
make it simple to exec kratos commands such as "kratos proto add api/helloworld/demo.proto".

# install

```bash
go install github.com/orzkratos/clikratos/cmd/clikratos-add-proto@latest
```

# command

new demo proto:
```bash
cd project-path/api/helloworld && clikratos-add-proto demo.proto
```

same with this:
```bash
cd project-path && kratos proto add api/helloworld/demo.proto
```

simple command:
```bash
cd project-path/api/helloworld

clikratos-add-proto demo
```
