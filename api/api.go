//go:generate mkdir -p timetrackapi/v1
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=timetrack/v1/oapi-codegen.config.yaml timetrack/v1/openapi.yaml
//go:generate mkdir -p peopleinfoapi/v1
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=peopleinfo/v1/oapi-codegen.config.yaml peopleinfo/v1/openapi.yaml

package api
