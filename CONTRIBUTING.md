# Contributing

## Generating protobuf code for Go

```
# First CD into the "protos" directory and run the below command
protoc -I=. --go_out=.. *.proto
```