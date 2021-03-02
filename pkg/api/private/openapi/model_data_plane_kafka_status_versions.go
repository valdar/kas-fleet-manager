/*
 * Kafka Service Fleet Manager
 *
 * Kafka Service Fleet Manager is a Rest API to manage kafka instances and connectors.
 *
 * API version: 0.0.1
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// DataPlaneKafkaStatusVersions Version information related to a Kafka cluster
type DataPlaneKafkaStatusVersions struct {
	Kafka   string `json:"kafka,omitempty"`
	Strimzi string `json:"strimzi,omitempty"`
}