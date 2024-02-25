// Package gen provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.16.2 DO NOT EDIT.
package gen

// Attachment defines model for Attachment.
type Attachment struct {
	Content string `json:"content"`
	Name    string `json:"name"`
}

// FastaRecord defines model for FastaRecord.
type FastaRecord struct {
	Identifier string `json:"identifier"`
	Sequence   string `json:"sequence"`
}

// PostExecuteLuaJSONBody defines parameters for PostExecuteLua.
type PostExecuteLuaJSONBody struct {
	Attachments *[]Attachment `json:"attachments,omitempty"`
	Script      string        `json:"script"`
}

// PostIoFastaParseTextBody defines parameters for PostIoFastaParse.
type PostIoFastaParseTextBody = string

// PostExecuteLuaJSONRequestBody defines body for PostExecuteLua for application/json ContentType.
type PostExecuteLuaJSONRequestBody PostExecuteLuaJSONBody

// PostIoFastaParseTextRequestBody defines body for PostIoFastaParse for text/plain ContentType.
type PostIoFastaParseTextRequestBody = PostIoFastaParseTextBody
