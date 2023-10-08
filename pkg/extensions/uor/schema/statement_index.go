package schema

import (
	ispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type Element struct {
	// LocatorType is the address of type information for the location.
	LocatorType string `json:"locatorType,omitempty"`
	// ResourceType is the address of type information for the resource.
	ResourceType string `json:"resourceType,omitempty"`
	// Location is the location of the resource.
	Location map[string]interface{} `json:"location,omitempty"`
	// Resource is the resource.
	Resource map[string]interface{} `json:"resource,omitempty"`
}

// Statement provides `application/vnd.uor.statement.v1+json` mediatype structure when marshalled to JSON.
type Statement struct {
	SchemaVersion int `json:"schemaVersion,omitempty"`

	// MediaType specifies the type of this document data structure e.g. `application/vnd.uor.statement.v1json`.
	ResType string `json:"resType,omitempty"`

	// Subject is the subject of the statement.
	Subject *Element `json:"subject,omitempty"`

	// Predicate is the predicate of the statement.
	Predicate *Element `json:"predicate,omitempty"`

	// Object is the object of the statement.
	Object *Element `json:"object,omitempty"`
}

type Location struct {
	Namespace string `json:"namespace,omitempty"`
}

/*
type Descriptor struct {
	// MediaType is the media type of the object this schema refers to.
	MediaType string `json:"mediaType,omitempty"`

	// Digest is the digest of the targeted content.
	Digest string `json:"digest"`

	// Size specifies the size in bytes of the blob.
	Size int `json:"size"`
}

type UORDescriptor struct {
	Descriptor
	Location
}*/

// StatementRecord is the database root. It is equivalent to an
// Element with a defined LocatorType of "uor_namespace" and
// a defined ResourceType of "uor_statement"
type StatementRecord struct {
	LocatorType  string           `json:"locatorType,omitempty"`
	Location     ispec.Descriptor `json:"location,omitempty"`
	ResourceType string           `json:"resourceType,omitempty"`
	Resource     Statement        `json:"resource,omitempty"`
}
