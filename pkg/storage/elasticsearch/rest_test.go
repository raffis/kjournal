package elasticsearch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	elasticsearch "github.com/elastic/go-elasticsearch/v8"
	"gotest.tools/v3/assert"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apiserver/pkg/registry/rest"
	srvstorage "k8s.io/apiserver/pkg/server/storage"
	"k8s.io/apiserver/pkg/storage/storagebackend"
)

// Mock transport replaces the HTTP transport for tests
type MockTransport struct {
	responseBody string
	middleware   func(req *http.Request, res *http.Response)
}

// RoundTrip returns a mock response.
func (t *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r := &http.Response{
		Body:   ioutil.NopCloser(strings.NewReader(t.responseBody)),
		Header: http.Header{"X-Elastic-Product": []string{"Elasticsearch"}},
	}

	if t.middleware != nil {
		t.middleware(req, r)
	}

	return r, nil
}

type listTest struct {
	name              string
	listOpts          func() *metainternalversion.ListOptions
	esResponse        esResults
	opts              Options
	expectedESRequest string
	expectedResult    *DummyList
	expectedError     error
}

func TestList(t *testing.T) {

	var tests = []listTest{
		{
			name: "Simple list request succeeds",
			listOpts: func() *metainternalversion.ListOptions {
				return &metainternalversion.ListOptions{
					LabelSelector: labels.Everything(),
					FieldSelector: fields.Everything(),
				}
			},
			esResponse: esResults{
				Hits: esHits{
					Hits: []esHit{
						{
							ID:     "a",
							Source: json.RawMessage(`{"field": "valueA"}`),
						},
					},
				},
			},
			expectedESRequest: `{"_source":{"excludes":["kind","apiVersion"]},"query":{"bool":{"must":[{"bool":{"should":null}},{"bool":{"should":[{"match_phrase":{"metadata.namespace":""}}]}}],"must_not":[]}},"sort":[{"metadata.uid":{"order":"asc","unmapped_type":"long"}}]}
`,
			expectedResult: &DummyList{
				Items: []Dummy{
					{
						ObjectMeta: v1.ObjectMeta{
							UID: "a",
						},
					},
				},
			},
		},
		{
			name: "Get a continue token if requested > total items",
			listOpts: func() *metainternalversion.ListOptions {
				return &metainternalversion.ListOptions{
					LabelSelector: labels.Everything(),
					FieldSelector: fields.Everything(),
					Limit:         1,
				}
			},
			esResponse: esResults{
				Hits: esHits{
					Hits: []esHit{
						{
							ID:     "a",
							Source: json.RawMessage(`{"field": "valueA"}`),
							Sort:   []interface{}{"sortFieldA"},
						},
					},
				},
			},
			expectedESRequest: `{"_source":{"excludes":["kind","apiVersion"]},"query":{"bool":{"must":[{"bool":{"should":null}},{"bool":{"should":[{"match_phrase":{"metadata.namespace":""}}]}}],"must_not":[]}},"sort":[{"metadata.uid":{"order":"asc","unmapped_type":"long"}}]}
`,
			expectedResult: &DummyList{
				ListMeta: v1.ListMeta{
					Continue: `["sortFieldA"]`,
				},
				Items: []Dummy{
					{
						ObjectMeta: v1.ObjectMeta{
							UID: "a",
						},
					},
				},
			},
		},
		{
			name: "Get next and last batch of items with a continue token",
			listOpts: func() *metainternalversion.ListOptions {
				return &metainternalversion.ListOptions{
					LabelSelector: labels.Everything(),
					FieldSelector: fields.Everything(),
					Limit:         1,
					Continue:      `["sortFieldA"]`,
				}
			},
			esResponse: esResults{
				Hits: esHits{
					Hits: []esHit{
						{
							ID:     "b",
							Source: json.RawMessage(`{"field": "valueB"}`),
						},
					},
				},
			},
			expectedESRequest: `{"_source":{"excludes":["kind","apiVersion"]},"query":{"bool":{"must":[{"bool":{"should":null}},{"bool":{"should":[{"match_phrase":{"metadata.namespace":""}}]}}],"must_not":[]}},"search_after":["sortFieldA"],"sort":[{"metadata.uid":{"order":"asc","unmapped_type":"long"}}]}
`,
			expectedResult: &DummyList{
				Items: []Dummy{
					{
						ObjectMeta: v1.ObjectMeta{
							UID: "b",
						},
					},
				},
			},
		},
		{
			name: "An invalid continue token ends with an error",
			listOpts: func() *metainternalversion.ListOptions {
				return &metainternalversion.ListOptions{
					LabelSelector: labels.Everything(),
					FieldSelector: fields.Everything(),
					Limit:         1,
					Continue:      `{"Invalid json`,
				}
			},
			esResponse: esResults{
				Hits: esHits{
					Hits: []esHit{
						{
							ID:     "b",
							Source: json.RawMessage(`{"field": "valueB"}`),
						},
					},
				},
			},
			expectedError: errors.New("failed to decode continue token: unexpected end of JSON input"),
		},
		{
			name: "All field selectors get mapped to a correct elasticsearch query",
			listOpts: func() *metainternalversion.ListOptions {
				selectors, _ := labels.Parse("fieldA<1,fieldA>1,fieldA=fieldB,fieldA==fieldB,fieldA!=fieldB,!fieldA,fieldA")

				return &metainternalversion.ListOptions{
					LabelSelector: selectors,
					FieldSelector: fields.Everything(),
				}
			},
			expectedESRequest: `{"_source":{"excludes":["kind","apiVersion"]},"query":{"bool":{"must":[{"bool":{"should":[{"range":{"fieldA":{"lt":"1"}}}]}},{"bool":{"should":[{"range":{"fieldA":{"gt":"1"}}}]}},{"bool":{"should":[{"match_phrase":{"fieldA":"fieldB"}}]}},{"bool":{"should":[{"match_phrase":{"fieldA":"fieldB"}}]}},{"bool":{"should":[null]}},{"bool":{"should":null}},{"bool":{"should":[{"match_phrase":{"metadata.namespace":""}}]}}],"must_not":[{"bool":{"should":[{"match_phrase":{"fieldA":"fieldB"}}]}},{"bool":{"should":[{"exists":{"field":"fieldA"}}]}}]}},"sort":[{"metadata.uid":{"order":"asc","unmapped_type":"long"}}]}
`,
		},
		{
			name: "All field selectors get mapped to a correct elasticsearch query and a field map",
			opts: Options{
				FieldMap: map[string][]string{
					"fieldA": []string{"toFieldA"},
				},
			},
			listOpts: func() *metainternalversion.ListOptions {
				selectors, _ := labels.Parse("fieldA<1,fieldA>1,fieldA=fieldB,fieldA==fieldB,fieldA!=fieldB,!fieldA,fieldA")

				return &metainternalversion.ListOptions{
					LabelSelector: selectors,
					FieldSelector: fields.Everything(),
				}
			},
			expectedESRequest: `{"_source":{"excludes":["kind","apiVersion"]},"query":{"bool":{"must":[{"bool":{"should":[{"range":{"toFieldA":{"lt":"1"}}}]}},{"bool":{"should":[{"range":{"toFieldA":{"gt":"1"}}}]}},{"bool":{"should":[{"match_phrase":{"toFieldA":"fieldB"}}]}},{"bool":{"should":[{"match_phrase":{"toFieldA":"fieldB"}}]}},{"bool":{"should":[null]}},{"bool":{"should":null}},{"bool":{"should":[{"match_phrase":{"metadata.namespace":""}}]}}],"must_not":[{"bool":{"should":[{"match_phrase":{"toFieldA":"fieldB"}}]}},{"bool":{"should":[{"exists":{"field":"toFieldA"}}]}}]}},"sort":[{"metadata.uid":{"order":"asc","unmapped_type":"long"}}]}
`,
		},
		{
			name: "Query includes timestamp field in sort and query range with default",
			opts: Options{
				FieldMap: map[string][]string{
					"fieldA": []string{"toFieldA"},
				},
				DefaultTimeRange: "now-24h",
				Backend: OptionsBackend{
					TimestampFields: []string{"timestampField"},
				},
			},
			listOpts: func() *metainternalversion.ListOptions {
				//selectors, _ := labels.Parse("timestampField<1")

				return &metainternalversion.ListOptions{
					LabelSelector: labels.Everything(),
					FieldSelector: fields.Everything(),
				}
			},
			expectedESRequest: `{"_source":{"excludes":["kind","apiVersion"]},"query":{"bool":{"must":[{"bool":{"should":[{"range":{"timestampField":{"gte":"now-24h"}}}]}},{"bool":{"should":[{"match_phrase":{"metadata.namespace":""}}]}}],"must_not":[]}},"sort":[{"timestampField":{"order":"asc","unmapped_type":"long"}},{"metadata.uid":{"order":"asc","unmapped_type":"long"}}]}
`,
		},
		{
			name: "Default time range gets ignored if a filter matches the defined timestamp field",
			opts: Options{
				DefaultTimeRange: "now-24h",
				Backend: OptionsBackend{
					TimestampFields: []string{"timestampField"},
				},
			},
			listOpts: func() *metainternalversion.ListOptions {
				selectors, _ := labels.Parse("timestampField<1")

				return &metainternalversion.ListOptions{
					LabelSelector: selectors,
					FieldSelector: fields.Everything(),
				}
			},
			expectedESRequest: `{"_source":{"excludes":["kind","apiVersion"]},"query":{"bool":{"must":[{"bool":{"should":[{"range":{"timestampField":{"lt":"1"}}}]}},{"bool":{"should":[{"match_phrase":{"metadata.namespace":""}}]}}],"must_not":[]}},"sort":[{"timestampField":{"order":"asc","unmapped_type":"long"}},{"metadata.uid":{"order":"asc","unmapped_type":"long"}}]}
`,
		},
		{
			name: "Response gets mapped with a field map",
			opts: Options{
				FieldMap: map[string][]string{
					"payload.toFieldA":           []string{"fieldA"},
					"metadata.creationTimestamp": []string{"timestampField"},
				},
				DefaultTimeRange: "now-24h",
				Backend: OptionsBackend{
					TimestampFields: []string{"timestampField"},
				},
			},
			listOpts: func() *metainternalversion.ListOptions {
				return &metainternalversion.ListOptions{
					LabelSelector: labels.Everything(),
					FieldSelector: fields.Everything(),
				}
			},
			esResponse: esResults{
				Hits: esHits{
					Hits: []esHit{
						{
							ID:     "a",
							Source: json.RawMessage(`{"fieldA": "valueA", "timestampField":"2022-10-17T07:20:59+00:00"}`),
						},
					},
				},
			},
			expectedESRequest: `{"_source":{"excludes":["kind","apiVersion"]},"query":{"bool":{"must":[{"bool":{"should":[{"range":{"timestampField":{"gte":"now-24h"}}}]}},{"bool":{"should":[{"match_phrase":{"metadata.namespace":""}}]}}],"must_not":[]}},"sort":[{"timestampField":{"order":"asc","unmapped_type":"long"}},{"metadata.uid":{"order":"asc","unmapped_type":"long"}}]}
`,
			expectedResult: &DummyList{
				Items: []Dummy{
					{
						ObjectMeta: v1.ObjectMeta{
							UID: "a",
							CreationTimestamp: v1.Time{
								Time: time.Unix(1665991259, 0),
							},
						},
						Payload: json.RawMessage(`{"toFieldA":"valueA"}`),
					},
				},
			},
		},
	}

	/*

	   type pitStream struct {
	   	rest    *elasticsearchREST
	   	options *metainternalversion.ListOptions
	   	context context.Context
	   }

	   var _ rest.ResourceStreamer = &pitStream{}

	   func (obj *pitStream) GetObjectKind() schema.ObjectKind {
	   	return schema.EmptyObjectKind
	   }

	   func (obj *pitStream) DeepCopyObject() runtime.Object {
	   	panic("rest.PITStream does not implement DeepCopyObject")
	   }

	   func (s *pitStream) InputStream(ctx context.Context, apiVersion, acceptHeader string) (io.ReadCloser, bool, string, error) {
	   	stream := &stream{
	   		usePIT:      true,
	   		refreshRate: 0,
	   		rest:        s.rest,
	   		ch:          make(chan watch.Event, int(s.rest.opts.Backend.BulkSize)),
	   		done:        make(chan bool),
	   	}

	   	r := &streamReader{
	   		stream: stream,
	   	}

	   	go func() {
	   		s.options.Limit = s.rest.opts.Backend.BulkSize
	   		stream.Start(s.context, s.options)
	   	}()

	   	return r, false, "application/json", nil
	   }

	   type streamReader struct {
	   	stream *stream
	   }

	   func (r *streamReader) Read(dst []byte) (n int, err error) {
	   	read := func(doc watch.Event) (int, error) {
	   		b, err := json.Marshal(metav1.WatchEvent{
	   			Type:   string(doc.Type),
	   			Object: runtime.RawExtension{Object: doc.Object},
	   		})

	   		s := copy(dst, b)
	   		return s, err
	   	}

	   	select {
	   	case doc, ok := <-r.stream.ResultChan():
	   		if ok {
	   			return read(doc)
	   		}

	   		return 0, io.EOF
	   	case <-r.stream.done:
	   		close(r.stream.ch)
	   	}

	   	return 0, nil
	   }

	   func (r *streamReader) Close() error {
	   	r.stream.Stop()
	   	return nil
	   }

	*/
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			responseBody, err := json.Marshal(test.esResponse)
			assert.NilError(t, err)

			transport := &MockTransport{
				middleware: func(req *http.Request, res *http.Response) {
					reqBody, err := io.ReadAll(req.Body)
					assert.NilError(t, err)
					assert.Equal(t, test.expectedESRequest, string(reqBody))

					fmt.Printf("body: %#v\n", string(reqBody))
				},
				responseBody: string(responseBody),
			}

			client, _ := elasticsearch.NewClient(elasticsearch.Config{Transport: transport})
			dummy := &Dummy{}
			scheme := &runtime.Scheme{}

			codec, _, _ := srvstorage.NewStorageCodec(srvstorage.StorageCodecConfig{
				StorageMediaType:  runtime.ContentTypeJSON,
				StorageSerializer: serializer.NewCodecFactory(scheme),
				//StorageVersion:    scheme.PrioritizedVersionsForGroup(dummy.GetGroupVersionResource().Group)[0],
				//(MemoryVersion:     scheme.PrioritizedVersionsForGroup(dummy.GetGroupVersionResource().Group)[0],
				Config: storagebackend.Config{},
			})

			restStorage := NewElasticsearchREST(
				dummy.GetGroupVersionResource().GroupResource(),
				codec,
				client,
				test.opts,
				dummy.NamespaceScoped(),
				dummy.New,
				dummy.NewList,
			)

			list, err := restStorage.(rest.Lister).List(context.TODO(), test.listOpts())

			if test.expectedError != nil {
				assert.Error(t, err, test.expectedError.Error())
			} else {
				assert.NilError(t, err)
				dummyList := list.(*DummyList)

				if test.expectedResult != nil {
					assert.Equal(t, test.expectedResult.Continue, dummyList.Continue)
					assert.Equal(t, len(test.expectedResult.Items), len(dummyList.Items))
					for k, dummy := range dummyList.Items {
						assert.Equal(t, test.expectedResult.Items[k].UID, dummy.UID)
						assert.Equal(t, test.expectedResult.Items[k].CreationTimestamp, dummy.CreationTimestamp)
						assert.Equal(t, string(test.expectedResult.Items[k].Payload), string(dummy.Payload))
					}
				}
			}
		})
	}
}
