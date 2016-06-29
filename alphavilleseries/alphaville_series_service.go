package alphavilleseries

import (
	"encoding/json"

	"github.com/Financial-Times/neo-utils-go/neoutils"
	"github.com/jmcvetta/neoism"
)

type service struct {
	cypherRunner neoutils.CypherRunner
	indexManager neoutils.IndexManager
}

// NewCypherAlphavilleSeriesService provides functions for create, update, delete operations on alphavilleSeries in Neo4j,
// plus other utility functions needed for a service
func NewCypherAlphavilleSeriesService(cypherRunner neoutils.CypherRunner, indexManager neoutils.IndexManager) service {
	return service{cypherRunner, indexManager}
}

func (s service) Initialise() error {
	return neoutils.EnsureConstraints(s.indexManager, map[string]string{
		"Thing":            "uuid",
		"Concept":          "uuid",
		"Classification":   "uuid",
		"AlphavilleSeries": "uuid"})
}

func (s service) Read(uuid string) (interface{}, bool, error) {
	results := []AlphavilleSeries{}

	query := &neoism.CypherQuery{
		Statement: `MATCH (n:AlphavilleSeries {uuid:{uuid}}) return n.uuid as uuid,
		n.prefLabel as prefLabel,
		n.description as description,
		n.tmeIdentifier as tmeIdentifier`,
		Parameters: map[string]interface{}{
			"uuid": uuid,
		},
		Result: &results,
	}

	err := s.cypherRunner.CypherBatch([]*neoism.CypherQuery{query})

	if err != nil {
		return AlphavilleSeries{}, false, err
	}

	if len(results) == 0 {
		return AlphavilleSeries{}, false, nil
	}

	return results[0], true, nil
}

func (s service) Write(thing interface{}) error {

	alphavilleSeries := thing.(AlphavilleSeries)

	params := map[string]interface{}{
		"uuid": alphavilleSeries.UUID,
	}

	if alphavilleSeries.PrefLabel != "" {
		params["prefLabel"] = alphavilleSeries.PrefLabel
	}

	if alphavilleSeries.TmeIdentifier != "" {
		params["tmeIdentifier"] = alphavilleSeries.TmeIdentifier
	}

	query := &neoism.CypherQuery{
		Statement: `MERGE (n:Thing {uuid: {uuid}})
					set n={allprops}
					set n :Concept
					set n :Classification
					set n :AlphavilleSeries
		`,
		Parameters: map[string]interface{}{
			"uuid":     alphavilleSeries.UUID,
			"allprops": params,
		},
	}

	return s.cypherRunner.CypherBatch([]*neoism.CypherQuery{query})
}

func (s service) Delete(uuid string) (bool, error) {
	clearNode := &neoism.CypherQuery{
		Statement: `
			MATCH (s:Thing {uuid: {uuid}})
			REMOVE s:Concept
			REMOVE s:Classification
			REMOVE s:AlphavilleSeries
			SET s={props}
		`,
		Parameters: map[string]interface{}{
			"uuid": uuid,
			"props": map[string]interface{}{
				"uuid": uuid,
			},
		},
		IncludeStats: true,
	}

	removeNodeIfUnused := &neoism.CypherQuery{
		Statement: `
			MATCH (s:Thing {uuid: {uuid}})
			OPTIONAL MATCH (s)-[a]-(x)
			WITH s, count(a) AS relCount
			WHERE relCount = 0
			DELETE s
		`,
		Parameters: map[string]interface{}{
			"uuid": uuid,
		},
	}

	err := s.cypherRunner.CypherBatch([]*neoism.CypherQuery{clearNode, removeNodeIfUnused})

	s1, err := clearNode.Stats()
	if err != nil {
		return false, err
	}

	var deleted bool
	if s1.ContainsUpdates && s1.LabelsRemoved > 0 {
		deleted = true
	}

	return deleted, err
}

func (s service) DecodeJSON(dec *json.Decoder) (interface{}, string, error) {
	alphavilleSeries := AlphavilleSeries{}
	err := dec.Decode(&alphavilleSeries)
	return alphavilleSeries, alphavilleSeries.UUID, err
}

func (s service) Check() error {
	return neoutils.Check(s.cypherRunner)
}

func (s service) Count() (int, error) {
	results := []struct {
		Count int `json:"c"`
	}{}

	query := &neoism.CypherQuery{
		Statement: `MATCH (n:AlphavilleSeries) return count(n) as c`,
		Result:    &results,
	}

	err := s.cypherRunner.CypherBatch([]*neoism.CypherQuery{query})

	if err != nil {
		return 0, err
	}

	return results[0].Count, nil
}