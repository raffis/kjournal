package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	"k8s.io/apimachinery/pkg/labels"
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

type queryBuilderFunc func() error
type queryBuilder struct {
	ctx     context.Context
	options *metainternalversion.ListOptions
	rest    *elasticsearchREST
	query   map[string]interface{}
}

func queryFromListOptions(ctx context.Context, options *metainternalversion.ListOptions, rest *elasticsearchREST) (map[string]interface{}, error) {
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

	req, _ := options.LabelSelector.Requirements()

	builders := []queryBuilderFunc{
		q.continueToken,
		q.sortByTimestampFields,
		q.fieldSelectors(req),
		q.fieldSelectors(rest.opts.Filter),
		q.defaultRange,
		q.namespaceFilter,
	}

	for _, builder := range builders {
		if err := builder(); err != nil {
			return q.query, err
		}
	}

	return q.query, nil
}

func (b *queryBuilder) fieldMapping(field string, defaultMap []string) []string {
	if val, ok := b.rest.opts.FieldMap[field]; ok {
		return val
	}

	return defaultMap
}

func (b *queryBuilder) continueToken() error {
	if b.options.Continue == "" {
		return nil
	}

	var searchAfter []interface{}
	err := json.Unmarshal([]byte(b.options.Continue), &searchAfter)
	if err != nil {
		return fmt.Errorf("failed to decode continue token: %w", err)
	}

	b.query["search_after"] = searchAfter
	return nil
}

func (b *queryBuilder) sortByTimestampFields() error {
	for _, tsField := range b.rest.opts.Backend.TimestampFields {
		b.query["sort"] = append(b.query["sort"].([]map[string]interface{}), map[string]interface{}{
			tsField: map[string]interface{}{
				"order":         "asc",
				"unmapped_type": "long",
			},
		})
	}

	for _, uidField := range b.fieldMapping("metadata.uid", []string{}) {
		b.query["sort"] = append(b.query["sort"].([]map[string]interface{}), map[string]interface{}{
			uidField: map[string]interface{}{
				"order":         "asc",
				"unmapped_type": "long",
			},
		})
	}

	return nil
}

func (b *queryBuilder) fieldSelectors(requirements labels.Requirements) queryBuilderFunc {
	return func() error {
		for _, req := range requirements {
			operator, ok := operatorMap[req.Operator()]
			if !ok {
				return fmt.Errorf("invalid selector operator %s", operator)
			}

			q := b.query["query"].(map[string]interface{})["bool"].(map[string]interface{})[operator[0]].([]map[string]interface{})
			fieldsMap := []string{req.Key()}

			for field, fieldsTo := range b.rest.opts.FieldMap {
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
			}

			q = append(q, map[string]interface{}{
				"bool": map[string]interface{}{
					"should": should,
				},
			})

			b.query["query"].(map[string]interface{})["bool"].(map[string]interface{})[operator[0]] = q
		}

		return nil
	}
}

func (b *queryBuilder) defaultRange() error {
	var skipTimestampFilter bool
	requirements, _ := b.options.LabelSelector.Requirements()

	for _, req := range requirements {
		operator, ok := operatorMap[req.Operator()]
		if !ok {
			return fmt.Errorf("invalid selector operator %s", operator)
		}

		fieldsMap := []string{req.Key()}

		for field, fieldsTo := range b.rest.opts.FieldMap {
			for k, fieldTo := range fieldsTo {
				lookupKey := strings.TrimLeft(strings.Replace(req.Key(), field, fieldTo, -1), ".")
				if lookupKey != req.Key() {
					fieldsMap[k] = lookupKey
					break
				}
			}
		}

		for _, fieldTo := range fieldsMap {
			for _, tsField := range b.rest.opts.Backend.TimestampFields {
				if !skipTimestampFilter && fieldTo == tsField {
					skipTimestampFilter = true
				}
			}
		}
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

	return nil
}

func (b *queryBuilder) namespaceFilter() error {
	if !b.rest.isNamespaced {
		return nil
	}

	ns, _ := request.NamespaceFrom(b.ctx)
	nsFields := b.fieldMapping("metadata.namespace", []string{"metadata.namespace"})
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
	return nil
}
