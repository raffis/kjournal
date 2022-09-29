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
	for _, fieldMap := range b.rest.opts.FieldMap {
		if fieldMap.Field == field {
			return fieldMap.Lookup
		}
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
	for _, tsField := range b.rest.opts.Backend.TimestampFields {
		b.query["sort"] = append(b.query["sort"].([]map[string]interface{}), map[string]interface{}{
			tsField: map[string]interface{}{
				"order":         "asc",
				"unmapped_type": "long",
			},
		})
	}

	return b
}

func (b *queryBuilder) fieldSelectors() *queryBuilder {
	var skipTimestampFilter bool
	requirements, _ := b.options.LabelSelector.Requirements()

	for _, req := range requirements {
		operator := operatorMap[req.Operator()]

		q := b.query["query"].(map[string]interface{})["bool"].(map[string]interface{})[operator[0]].([]map[string]interface{})
		fieldsMap := []string{req.Key()}

		for _, fieldMap := range b.rest.opts.FieldMap {
			for k, fieldTo := range fieldMap.Lookup {
				lookupKey := strings.TrimLeft(strings.Replace(req.Key(), fieldMap.Field, fieldTo, -1), ".")
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

			for _, tsField := range b.rest.opts.Backend.TimestampFields {
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

		for _, tsField := range b.rest.opts.Backend.TimestampFields {
			should = append(should, map[string]interface{}{
				"range": map[string]interface{}{
					tsField: map[string]interface{}{
						"gte": b.rest.opts.DefaultTimeRange,
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
