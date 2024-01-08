// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/servers/broadcastwake": {
            "post": {
                "description": "Wake a server using Wake on LAN",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "servers"
                ],
                "summary": "Wake a server using just a mac (broadcast is enumerated)",
                "parameters": [
                    {
                        "description": "Broadcast wake request",
                        "name": "broadcastWakeRequest",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handlers.BroadcastWakeRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Wake on LAN packets successfully sent to all available broadcast addresses",
                        "schema": {
                            "$ref": "#/definitions/handlers.Response"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/handlers.Response"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/handlers.Response"
                        }
                    }
                }
            }
        },
        "/servers/wake": {
            "post": {
                "description": "Wake a server using Wake on LAN",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "servers"
                ],
                "summary": "Wake a server using a mac and a broadcast address",
                "parameters": [
                    {
                        "description": "Wake server request",
                        "name": "wakeServerRequest",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handlers.WakeServerRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/handlers.Response"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/handlers.Response"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/handlers.Response"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "handlers.BroadcastWakeRequest": {
            "type": "object",
            "required": [
                "mac"
            ],
            "properties": {
                "mac": {
                    "type": "string",
                    "example": "00:11:22:33:44:55"
                },
                "port": {
                    "type": "integer",
                    "maximum": 65535,
                    "minimum": 1,
                    "example": 9
                }
            }
        },
        "handlers.Response": {
            "type": "object",
            "properties": {
                "message": {
                    "type": "string"
                }
            }
        },
        "handlers.WakeServerRequest": {
            "type": "object",
            "required": [
                "broadcast",
                "mac"
            ],
            "properties": {
                "broadcast": {
                    "type": "string",
                    "example": "192.168.1.13"
                },
                "mac": {
                    "type": "string",
                    "example": "00:11:22:33:44:55"
                },
                "port": {
                    "type": "integer",
                    "maximum": 65535,
                    "minimum": 1,
                    "example": 9
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "",
	Description:      "",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}