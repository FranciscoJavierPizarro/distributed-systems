CGO_ENABLED=0 go build -o . ./cmd/clienttest/...
mv clienttest client
cd client
docker build . -t localhost:5001/clienttest:latest
docker push localhost:5001/clienttest:latest
