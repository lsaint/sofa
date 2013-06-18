protoc --go_out=. proto/proto.proto
go build -race sofa.go
