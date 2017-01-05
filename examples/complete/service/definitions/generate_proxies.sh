# regenerate proxies for protocol buffer definitions
protoc -I protos protos/sum/sum.proto --go_out=plugins=grpc:proxies
