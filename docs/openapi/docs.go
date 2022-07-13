// Package openapi GENERATED BY SWAG; DO NOT EDIT
// This file was generated by swaggo/swag
package openapi

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "termsOfService": "https://www.bytebase.com/terms",
        "contact": {
            "name": "API Support",
            "url": "https://github.com/youzi-1122/bytebase/",
            "email": "support@bytebase.com"
        },
        "license": {
            "name": "MIT",
            "url": "https://github.com/youzi-1122/bytebase/blob/main/LICENSE"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/sql/advise": {
            "get": {
                "description": "Parse and check the SQL statement according to the schema review policy.",
                "consumes": [
                    "*/*"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Schema Review"
                ],
                "summary": "Check the SQL statement.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "The environment name. Case sensitive.",
                        "name": "environment",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "The SQL statement.",
                        "name": "statement",
                        "in": "query",
                        "required": true
                    },
                    {
                        "enum": [
                            "MySQL",
                            "PostgreSQL",
                            "TiDB"
                        ],
                        "type": "string",
                        "description": "The database type. Required if the port, host and database name is not specified.",
                        "name": "databaseType",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "The instance host.",
                        "name": "host",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "The instance port.",
                        "name": "port",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "The database name in the instance.",
                        "name": "databaseName",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/advisor.Advice"
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/echo.HTTPError"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/echo.HTTPError"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "advisor.Advice": {
            "type": "object",
            "properties": {
                "code": {
                    "description": "Code is the SQL check error code.",
                    "type": "integer"
                },
                "content": {
                    "type": "string"
                },
                "status": {
                    "description": "Status is the SQL check result. Could be \"SUCCESS\", \"WARN\", \"ERROR\"",
                    "type": "string"
                },
                "title": {
                    "type": "string"
                }
            }
        },
        "echo.HTTPError": {
            "type": "object",
            "properties": {
                "message": {}
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:8080",
	BasePath:         "/v1/",
	Schemes:          []string{"http"},
	Title:            "Bytebase OpenAPI",
	Description:      "The OpenAPI for bytebase.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
