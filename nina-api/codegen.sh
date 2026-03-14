function generate_client() {
    go get -tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest 
    wget https://raw.githubusercontent.com/christian-photo/ninaAPI/refs/heads/dev/ninaAPI/api_spec.yaml -O ninaAPI.yaml > /dev/null 2>&1
    go get -u github.com/oapi-codegen/runtime
    go generate ./...
    rm ninaAPI.yaml
}

generate_client