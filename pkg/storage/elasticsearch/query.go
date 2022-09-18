package elasticsearch

import (
	"context"
	"encoding/json"
	"strings"

	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apiserver/pkg/endpoints/request"
)

var operatorMap = map[selection.Operator][]string{
	selection.Equals:       {"must", "match_phrase"},
	selection.DoubleEquals: {"must", "match_phrase"},
	selection.NotEquals:    {"must_not", "match_phrase"},
	selection.GreaterThan:  {"must", "range"},
	selection.LessThan:     {"must", "range"},
	selection.DoesNotExist: {"must_not", "exists"},
	selection.Exists:       {"must", "exists"},
}

type queryBuilder struct {
	ctx     context.Context
	options *metainternalversion.ListOptions
	rest    *elasticsearchREST
	query   map[string]interface{}
}

func QueryFromListOptions(ctx context.Context, options *metainternalversion.ListOptions, rest *elasticsearchREST) map[string]interface{} {
	q := queryBuilder{
		rest:    rest,
		ctx:     ctx,
		options: options,
		query: map[string]interface{}{
			"_source": map[string]interface{}{
				"excludes": []interface{}{"kind", "apiVersion"},
			},
			"sort": []map[string]interface{}{},
			"query": map[string]interface{}{
				"bool": map[string]interface{}{
					"must":     []map[string]interface{}{},
					"must_not": []map[string]interface{}{},
				},
			},
		},
	}

	q.continueToken().
		sortByTimestampFields().
		fieldSelectors().
		namespaceFilter()

	return q.query
}

func (b *queryBuilder) fieldMapping(field string) []string {
	if val, ok := b.rest.apiBinding.FieldMap[field]; ok {
		return val
	}

	return []string{field}
}

func (b *queryBuilder) continueToken() *queryBuilder {
	if b.options.Continue == "" {
		return b
	}

	var searchAfter []interface{}
	err := json.Unmarshal([]byte(b.options.Continue), &searchAfter)
	if err != nil {
		return b
	}

	b.query["search_after"] = searchAfter
	return b
}

func (b *queryBuilder) sortByTimestampFields() *queryBuilder {
	tsFields := b.fieldMapping("metadata.creationTimestamp")

	for _, tsField := range tsFields {
		b.query["sort"] = append(b.query["sort"].([]map[string]interface{}), map[string]interface{}{
			tsField: "asc",
		})
	}

	return b
}

func (b *queryBuilder) fieldSelectors() *queryBuilder {
	tsFields := b.fieldMapping("metadata.creationTimestamp")
	var skipTimestampFilter bool
	requirements, _ := b.options.LabelSelector.Requirements()

	for _, req := range requirements {
		operator := operatorMap[req.Operator()]

		q := b.query["query"].(map[string]interface{})["bool"].(map[string]interface{})[operator[0]].([]map[string]interface{})
		fieldsMap := []string{req.Key()}

		for field, fieldsTo := range b.rest.apiBinding.FieldMap {
			for k, fieldTo := range fieldsTo {
				lookupKey := strings.TrimLeft(strings.Replace(req.Key(), field, fieldTo, -1), ".")
				if lookupKey != req.Key() {
					fieldsMap[k] = lookupKey
					break
				}
			}
		}

		var should []map[string]interface{}
		for _, fieldTo := range fieldsMap {
			var shouldCondition map[string]interface{}
			switch req.Operator() {
			case selection.LessThan:
				shouldCondition = map[string]interface{}{
					operator[1]: map[string]interface{}{
						fieldTo: map[string]interface{}{
							"lt": req.Values().UnsortedList()[0],
						},
					},
				}
			case selection.GreaterThan:
				shouldCondition = map[string]interface{}{
					operator[1]: map[string]interface{}{
						fieldTo: map[string]interface{}{
							"gt": req.Values().UnsortedList()[0],
						},
					},
				}
			case selection.Exists:
			case selection.DoesNotExist:
				shouldCondition = map[string]interface{}{
					operator[1]: map[string]interface{}{
						"field": fieldTo,
					},
				}
			default:
				shouldCondition = map[string]interface{}{
					operator[1]: map[string]interface{}{
						fieldTo: req.Values().UnsortedList()[0],
					},
				}
			}

			should = append(should, shouldCondition)

			for _, tsField := range tsFields {
				if !skipTimestampFilter && fieldTo == tsField {
					skipTimestampFilter = true
				}
			}
		}

		q = append(q, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": should,
			},
		})

		b.query["query"].(map[string]interface{})["bool"].(map[string]interface{})[operator[0]] = q
	}

	if !skipTimestampFilter {
		q := b.query["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"].([]map[string]interface{})
		var should []map[string]interface{}

		for _, tsField := range tsFields {
			should = append(should, map[string]interface{}{
				"range": map[string]interface{}{
					tsField: map[string]interface{}{
						"gte": "now-5h",
					},
				},
			})
		}

		q = append(q, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": should,
			},
		})

		b.query["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"] = q

	}

	return b
}

func (b *queryBuilder) namespaceFilter() *queryBuilder {
	if !b.rest.isNamespaced {
		return b
	}

	ns, _ := request.NamespaceFrom(b.ctx)
	nsFields := b.fieldMapping("metadata.namespace")
	q := b.query["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"].([]map[string]interface{})
	var should []map[string]interface{}

	for _, nsField := range nsFields {
		should = append(should, map[string]interface{}{
			"match_phrase": map[string]interface{}{
				nsField: ns,
			},
		})
	}

	q = append(q, map[string]interface{}{
		"bool": map[string]interface{}{
			"should": should,
		},
	})

	b.query["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"] = q
	return b
}
