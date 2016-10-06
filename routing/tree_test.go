package routing

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/zalando/skipper/logging/loggingtest"
	"github.com/zalando/skipper/routing/testdataclient"
)

type closer interface {
	Close()
}

type pathTestRouting struct {
	*Routing
	closer
}

type pathSpecVariant struct {
	route  string
	params map[string]string
}

type pathSpecTest struct {
	considerTrailing, ignoreTrailing pathSpecVariant
}

type pathSpec struct {
	routes string
	tests  map[string]pathSpecTest
}

func (r *pathTestRouting) Close() {
	r.Routing.Close()
	r.closer.Close()
}

func createPathMatchRouting(routes string, opts MatchingOptions) (*pathTestRouting, error) {
	dc, err := testdataclient.NewDoc(routes)
	if err != nil {
		return nil, err
	}

	logger := loggingtest.New()
	rt := New(Options{MatchingOptions: opts, DataClients: []DataClient{dc}, Log: logger})
	if err := logger.WaitFor("route settings applied", time.Second); err != nil {
		return nil, err
	}

	return &pathTestRouting{rt, logger}, nil
}

func paramsEqual(left, right map[string]string) bool {
	if len(left) != len(right) {
		return false
	}

	for k, v := range left {
		if right[k] != v {
			return false
		}
	}

	return true
}

func testPathMatchVariant(t *testing.T, path string, v pathSpecVariant, rt *Routing) {
	r, p := rt.Route(&http.Request{URL: &url.URL{Path: path}})

	if r != nil && v.route == "" {
		t.Error(path, "found route when should not", r.Id)
		return
	}

	if r == nil && v.route != "" {
		t.Error(path, "failed to find route", v.route)
		return
	}

	if r != nil && r.Id != v.route {
		t.Error(path, "found invalid route", r.Id, v.route)
		return
	}

	if !paramsEqual(p, v.params) {
		t.Error(path, "failed to return the right params", p, v.params)
	}
}

func testPathMatchSpec(t *testing.T, s pathSpec) {
	routingConsiderTrailing, err := createPathMatchRouting(s.routes, MatchingOptionsNone)
	if err != nil {
		t.Error(s.routes, "failed to create routing", err)
		return
	}

	defer routingConsiderTrailing.Close()

	routingIgnoreTrailing, err := createPathMatchRouting(s.routes, IgnoreTrailingSlash)
	if err != nil {
		t.Error(s.routes, "failed to create routing", err)
		return
	}

	defer routingIgnoreTrailing.Close()

	for path, expect := range s.tests {
		testPathMatchVariant(t, path, expect.considerTrailing, routingConsiderTrailing.Routing)
		if t.Failed() {
			return
		}

		testPathMatchVariant(t, path, expect.ignoreTrailing, routingIgnoreTrailing.Routing)
		if t.Failed() {
			return
		}
	}
}

func testPathMatch(t *testing.T, s []pathSpec) {
	for _, si := range s {
		if testPathMatchSpec(t, si); t.Failed() {
			return
		}
	}
}

func TestSinglePathMatch(t *testing.T) {
	testPathMatch(t, []pathSpec{{
		/* considering as unspecified
			`route1: Path("") -> <shunt>;`,
			map[string]pathSpecTest{
				"": {
					considerTrailing: pathSpecVariant{"route1", nil},
					ignoreTrailing: pathSpecVariant{"route1", nil},
				},
				"/": {
					considerTrailing: pathSpecVariant{"route1", nil},
					ignoreTrailing: pathSpecVariant{"route1", nil},
				},
				"/foo": {
					considerTrailing: pathSpecVariant{"route1", nil},
					ignoreTrailing: pathSpecVariant{"route1", nil},
				},
			},
		}, {
		*/
		`route1: Path("/") -> <shunt>;`,
		map[string]pathSpecTest{
			"": {
				considerTrailing: pathSpecVariant{"route1", nil},
				ignoreTrailing:   pathSpecVariant{"route1", nil},
			},
			"/": {
				considerTrailing: pathSpecVariant{"route1", nil},
				ignoreTrailing:   pathSpecVariant{"route1", nil},
			},
			"/foo": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
		},
	}, {
		`route1: Path("/foo") -> <shunt>;`,
		map[string]pathSpecTest{
			"": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo": {
				considerTrailing: pathSpecVariant{"route1", nil},
				ignoreTrailing:   pathSpecVariant{"route1", nil},
			},
			"/foo/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"route1", nil},
			},
			"/foo/bar": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
		},
	}, {
		`route1: Path("/foo/") -> <shunt>;`,
		map[string]pathSpecTest{
			"": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"route1", nil},
			},
			"/foo/": {
				considerTrailing: pathSpecVariant{"route1", nil},
				ignoreTrailing:   pathSpecVariant{"route1", nil},
			},
			"/foo/bar": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
		},
	}, {
		`route1: Path("/foo/bar") -> <shunt>;`,
		map[string]pathSpecTest{
			"": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar": {
				considerTrailing: pathSpecVariant{"route1", nil},
				ignoreTrailing:   pathSpecVariant{"route1", nil},
			},
			"/foo/bar/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"route1", nil},
			},
			"/foo/bar/baz": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar/baz/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
		},
	}, {
		`route1: Path("/foo/bar/") -> <shunt>;`,
		map[string]pathSpecTest{
			"": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"route1", nil},
			},
			"/foo/bar/": {
				considerTrailing: pathSpecVariant{"route1", nil},
				ignoreTrailing:   pathSpecVariant{"route1", nil},
			},
			"/foo/bar/baz": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar/baz/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
		},
	}})
}

func TestSimpleWildcardMatch(t *testing.T) {
	testPathMatch(t, []pathSpec{{
		/* considering as unspecified
			`route1: Path(":name1") -> <shunt>;`,
			map[string]pathSpecTest{
				"": {
					considerTrailing: pathSpecVariant{"", nil},
					ignoreTrailing: pathSpecVariant{"", nil},
				},
				"/": {
					considerTrailing: pathSpecVariant{"", nil},
					ignoreTrailing: pathSpecVariant{"", nil},
				},
				"/foo": {
					considerTrailing: pathSpecVariant{"route1", map[string]string{"name1": "foo"}},
					ignoreTrailing: pathSpecVariant{"route1", map[string]string{"name1": "foo"}},
				},
				"/foo/": {
					considerTrailing: pathSpecVariant{"", nil},
					ignoreTrailing: pathSpecVariant{"route1", map[string]string{"name1": "foo"}},
				},
				"/foo/bar": {
					considerTrailing: pathSpecVariant{"", nil},
					ignoreTrailing: pathSpecVariant{"", nil},
				},
				"/foo/bar/": {
					considerTrailing: pathSpecVariant{"", nil},
					ignoreTrailing: pathSpecVariant{"", nil},
				},
			},
		}, {
			`route1: Path(":name1/") -> <shunt>;`,
			map[string]pathSpecTest{
				"": {
					considerTrailing: pathSpecVariant{"", nil},
					ignoreTrailing: pathSpecVariant{"", nil},
				},
				"/": {
					considerTrailing: pathSpecVariant{"", nil},
					ignoreTrailing: pathSpecVariant{"", nil},
				},
				"/foo": {
					considerTrailing: pathSpecVariant{"", nil},
					ignoreTrailing: pathSpecVariant{"route1", map[string]string{"name1": "foo"}},
				},
				"/foo/": {
					considerTrailing: pathSpecVariant{"route1", map[string]string{"name1": "foo"}},
					ignoreTrailing: pathSpecVariant{"route1", map[string]string{"name1": "foo"}},
				},
				"/foo/bar": {
					considerTrailing: pathSpecVariant{"", nil},
					ignoreTrailing: pathSpecVariant{"", nil},
				},
				"/foo/bar/": {
					considerTrailing: pathSpecVariant{"", nil},
					ignoreTrailing: pathSpecVariant{"", nil},
				},
			},
		}, {
			`route1: Path("/foo:name1") -> <shunt>;`,
			map[string]pathSpecTest{
				"": {
					considerTrailing: pathSpecVariant{"", nil},
					ignoreTrailing: pathSpecVariant{"", nil},
				},
				"/": {
					considerTrailing: pathSpecVariant{"", nil},
					ignoreTrailing: pathSpecVariant{"", nil},
				},
				"/foo": {
					considerTrailing: pathSpecVariant{"", nil},
					ignoreTrailing: pathSpecVariant{"", nil},
				},
				"/foo/": {
					considerTrailing: pathSpecVariant{"", nil},
					ignoreTrailing: pathSpecVariant{"", nil},
				},
				"/foo/bar": {
					considerTrailing: pathSpecVariant{"", nil},
					ignoreTrailing: pathSpecVariant{"", nil},
				},
				"/foo/bar/": {
					considerTrailing: pathSpecVariant{"", nil},
					ignoreTrailing: pathSpecVariant{"", nil},
				},
			},
		}, {
			`route1: Path("/foo:name1/") -> <shunt>;`,
			map[string]pathSpecTest{
				"": {
					considerTrailing: pathSpecVariant{"", nil},
					ignoreTrailing: pathSpecVariant{"", nil},
				},
				"/": {
					considerTrailing: pathSpecVariant{"", nil},
					ignoreTrailing: pathSpecVariant{"", nil},
				},
				"/foo": {
					considerTrailing: pathSpecVariant{"", nil},
					ignoreTrailing: pathSpecVariant{"", nil},
				},
				"/foo/": {
					considerTrailing: pathSpecVariant{"", nil},
					ignoreTrailing: pathSpecVariant{"", nil},
				},
				"/foo/bar": {
					considerTrailing: pathSpecVariant{"", nil},
					ignoreTrailing: pathSpecVariant{"", nil},
				},
				"/foo/bar/": {
					considerTrailing: pathSpecVariant{"", nil},
					ignoreTrailing: pathSpecVariant{"", nil},
				},
			},
		}, {
		*/
		`route1: Path("/:name1") -> <shunt>;`,
		map[string]pathSpecTest{
			"": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"name1": "foo"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "foo"}},
			},
			"/foo/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "foo"}},
			},
			"/foo/bar": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
		},
	}, {
		`route1: Path("/:name1/") -> <shunt>;`,
		map[string]pathSpecTest{
			"": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "foo"}},
			},
			"/foo/": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"name1": "foo"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "foo"}},
			},
			"/foo/bar": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
		},
	}, {
		`route1: Path("/:name1/bar") -> <shunt>;`,
		map[string]pathSpecTest{
			"": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"name1": "foo"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "foo"}},
			},
			"/foo/bar/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "foo"}},
			},
			"/foo/bar/baz": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar/baz/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
		},
	}, {
		`route1: Path("/:name1/bar/") -> <shunt>;`,
		map[string]pathSpecTest{
			"": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "foo"}},
			},
			"/foo/bar/": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"name1": "foo"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "foo"}},
			},
			"/foo/bar/baz": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar/baz/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
		},
	}, {
		`route1: Path("/:name1/bar/") -> <shunt>;`,
		map[string]pathSpecTest{
			"": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "foo"}},
			},
			"/foo/bar/": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"name1": "foo"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "foo"}},
			},
			"/foo/bar/baz": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar/baz/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
		},
	}, {
		`route1: Path("/foo/:name1") -> <shunt>;`,
		map[string]pathSpecTest{
			"": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"name1": "bar"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "bar"}},
			},
			"/foo/bar/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "bar"}},
			},
			"/foo/bar/baz": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar/baz/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
		},
	}, {
		`route1: Path("/foo/:name1/") -> <shunt>;`,
		map[string]pathSpecTest{
			"": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "bar"}},
			},
			"/foo/bar/": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"name1": "bar"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "bar"}},
			},
			"/foo/bar/baz": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar/baz/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
		},
	}, {
		`route1: Path("/foo/:name1/baz") -> <shunt>;`,
		map[string]pathSpecTest{
			"": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar/baz": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"name1": "bar"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "bar"}},
			},
			"/foo/bar/baz/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "bar"}},
			},
		},
	}, {
		`route1: Path("/foo/:name1/baz/") -> <shunt>;`,
		map[string]pathSpecTest{
			"": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar/baz": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "bar"}},
			},
			"/foo/bar/baz/": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"name1": "bar"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "bar"}},
			},
		},
	}, {
		`route1: Path("/:name1/:name2") -> <shunt>;`,
		map[string]pathSpecTest{
			"": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"name1": "foo", "name2": "bar"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "foo", "name2": "bar"}},
			},
			"/foo/bar/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "foo", "name2": "bar"}},
			},
			"/foo/bar/baz": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar/baz/": {
				ignoreTrailing: pathSpecVariant{"", nil},
			},
		},
	}, {
		`route1: Path("/:name1/:name2/") -> <shunt>;`,
		map[string]pathSpecTest{
			"": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "foo", "name2": "bar"}},
			},
			"/foo/bar/": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"name1": "foo", "name2": "bar"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "foo", "name2": "bar"}},
			},
			"/foo/bar/baz": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar/baz/": {
				ignoreTrailing: pathSpecVariant{"", nil},
			},
		},
	}, {
		`route1: Path("/foo/:name1/:name2") -> <shunt>;`,
		map[string]pathSpecTest{
			"": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar/baz": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"name1": "bar", "name2": "baz"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "bar", "name2": "baz"}},
			},
			"/foo/bar/baz/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "bar", "name2": "baz"}},
			},
		},
	}, {
		`route1: Path("/foo/:name1/:name2/") -> <shunt>;`,
		map[string]pathSpecTest{
			"": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar/baz": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "bar", "name2": "baz"}},
			},
			"/foo/bar/baz/": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"name1": "bar", "name2": "baz"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "bar", "name2": "baz"}},
			},
		},
	}, {
		`route1: Path("/:name1/bar/:name2") -> <shunt>;`,
		map[string]pathSpecTest{
			"": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar/baz": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"name1": "foo", "name2": "baz"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "foo", "name2": "baz"}},
			},
			"/foo/bar/baz/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "foo", "name2": "baz"}},
			},
		},
	}, {
		`route1: Path("/:name1/bar/:name2/") -> <shunt>;`,
		map[string]pathSpecTest{
			"": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar/baz": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "foo", "name2": "baz"}},
			},
			"/foo/bar/baz/": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"name1": "foo", "name2": "baz"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "foo", "name2": "baz"}},
			},
		},
	}, {
		`route1: Path("/:name1/:name2/baz") -> <shunt>;`,
		map[string]pathSpecTest{
			"": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar/baz": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"name1": "foo", "name2": "bar"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "foo", "name2": "bar"}},
			},
			"/foo/bar/baz/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "foo", "name2": "bar"}},
			},
		},
	}, {
		`route1: Path("/:name1/:name2/baz/") -> <shunt>;`,
		map[string]pathSpecTest{
			"": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar/baz": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "foo", "name2": "bar"}},
			},
			"/foo/bar/baz/": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"name1": "foo", "name2": "bar"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "foo", "name2": "bar"}},
			},
		},
	}})
}

func TestCatchallWildcardMatch(t *testing.T) {
	testPathMatch(t, []pathSpec{{
	/* considering as unspecified
		`route1: Path("*") -> <shunt>;`,
		map[string]pathSpecTest{
			"": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"", nil},
			},
			"/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"", nil},
			},
			"/foo": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"": "foo"}},
				ignoreTrailing: pathSpecVariant{"route1", map[string]string{"": "foo"}},
			},
			"/foo/": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"": "foo/"}},
				ignoreTrailing: pathSpecVariant{"route1", map[string]string{"": "foo"}},
			},
			"/foo/bar": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"": "foo/bar"}},
				ignoreTrailing: pathSpecVariant{"route1", map[string]string{"": "foo/bar"}},
			},
			"/foo/bar/": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"": "foo/bar/"}},
				ignoreTrailing: pathSpecVariant{"route1", map[string]string{"": "foo/bar"}},
			},
		},
	}, {
		`route1: Path("*name1") -> <shunt>;`,
		map[string]pathSpecTest{
			"": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"", nil},
			},
			"/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"", nil},
			},
			"/foo": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"name1": "/foo"}},
				ignoreTrailing: pathSpecVariant{"route1", map[string]string{"name1": "/foo"}},
			},
			"/foo/": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"name1": "/foo/"}},
				ignoreTrailing: pathSpecVariant{"route1", map[string]string{"name1": "/foo"}},
			},
			"/foo/bar": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"name1": "/foo/bar"}},
				ignoreTrailing: pathSpecVariant{"route1", map[string]string{"name1": "/foo/bar"}},
			},
			"/foo/bar/": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"name1": "/foo/bar/"}},
				ignoreTrailing: pathSpecVariant{"route1", map[string]string{"name1": "/foo/bar"}},
			},
		},
	}, {
		`route1: Path("/*") -> <shunt>;`,
		map[string]pathSpecTest{
			"": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"", nil},
			},
			"/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"", nil},
			},
			"/foo": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"": "foo"}},
				ignoreTrailing: pathSpecVariant{"route1", map[string]string{"": "foo"}},
			},
			"/foo/": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"": "foo/"}},
				ignoreTrailing: pathSpecVariant{"route1", map[string]string{"": "foo"}},
			},
			"/foo/bar": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"": "foo/bar"}},
				ignoreTrailing: pathSpecVariant{"route1", map[string]string{"": "foo/bar"}},
			},
			"/foo/bar/": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"": "foo/bar/"}},
				ignoreTrailing: pathSpecVariant{"route1", map[string]string{"": "foo/bar"}},
			},
		},
	}, {
		`route1: Path("/foo*") -> <shunt>;`,
		map[string]pathSpecTest{
			"": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"", nil},
			},
			"/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"", nil},
			},
			"/foo": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"", nil},
			},
			"/foo/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"", nil},
			},
			"/foo/bar": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"", nil},
			},
			"/foo/bar/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"", nil},
			},
		},
	}, {
		`route1: Path("/foo/*") -> <shunt>;`,
		map[string]pathSpecTest{
			"": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"", nil},
			},
			"/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"", nil},
			},
			"/foo": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"", nil},
			},
			"/foo/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"", nil},
			},
			"/foo/bar": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"": "bar"}},
				ignoreTrailing: pathSpecVariant{"route1", map[string]string{"": "bar"}},
			},
			"/foo/bar/": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"": "bar/"}},
				ignoreTrailing: pathSpecVariant{"route1", map[string]string{"": "bar"}},
			},
		},
	}, {
		`route1: Path("/foo*name") -> <shunt>;`,
		map[string]pathSpecTest{
			"": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"", nil},
			},
			"/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"", nil},
			},
			"/foo": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"", nil},
			},
			"/foo/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"", nil},
			},
			"/foo/bar": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"", nil},
			},
			"/foo/bar/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"", nil},
			},
		},
	}, {
		`route1: Path("/*name1/") -> <shunt>;`,
		map[string]pathSpecTest{
			"": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"", nil},
			},
			"/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"", nil},
			},
			"/foo": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"route1", map[string]string{"name1": "/foo"}},
			},
			"/foo/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"route1", map[string]string{"name1": "/foo"}},
			},
			"/foo/bar": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"route1", map[string]string{"name1": "/foo/bar"}},
			},
			"/foo/bar/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"route1", map[string]string{"name1": "/foo/bar"}},
			},
		},
	}, {
		`route1: Path("/*name1/bar") -> <shunt>;`,
		map[string]pathSpecTest{
			"": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"", nil},
			},
			"/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"", nil},
			},
			"/foo": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"", nil},
			},
			"/foo/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"", nil},
			},
			"/foo/bar": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"", nil},
			},
			"/foo/bar/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"", nil},
			},
		},
	}, {
		`route1: Path("/*name1/bar/") -> <shunt>;`,
		map[string]pathSpecTest{
			"": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"", nil},
			},
			"/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"", nil},
			},
			"/foo": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"", nil},
			},
			"/foo/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"", nil},
			},
			"/foo/bar": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"", nil},
			},
			"/foo/bar/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"", nil},
			},
		},
	}, {
		`route1: Path("/foo/*name1/") -> <shunt>;`,
		map[string]pathSpecTest{
			"": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"", nil},
			},
			"/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"", nil},
			},
			"/foo": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"", nil},
			},
			"/foo/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"", nil},
			},
			"/foo/bar": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"route1", map[string]string{"name1": "/bar"}},
			},
			"/foo/bar/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"route1", map[string]string{"name1": "/bar"}},
			},
			"/foo/bar/baz": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"route1", map[string]string{"name1": "/bar/baz"}},
			},
			"/foo/bar/baz/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing: pathSpecVariant{"route1", map[string]string{"name1": "/bar/baz"}},
			},
		},
	*/
	}, {
		`route1: Path("/*name1") -> <shunt>;`,
		map[string]pathSpecTest{
			"": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"name1": "/foo"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "/foo"}},
			},
			"/foo/": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"name1": "/foo/"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "/foo"}},
			},
			"/foo/bar": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"name1": "/foo/bar"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "/foo/bar"}},
			},
			"/foo/bar/": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"name1": "/foo/bar/"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "/foo/bar"}},
			},
			"/foo/bar/baz": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"name1": "/foo/bar/baz"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "/foo/bar/baz"}},
			},
			"/foo/bar/baz/": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"name1": "/foo/bar/baz/"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "/foo/bar/baz"}},
			},
		},
	}, {
		`route1: Path("/foo/*name1") -> <shunt>;`,
		map[string]pathSpecTest{
			"": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"name1": "/bar"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "/bar"}},
			},
			"/foo/bar/": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"name1": "/bar/"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "/bar"}},
			},
			"/foo/bar/baz": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"name1": "/bar/baz"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "/bar/baz"}},
			},
			"/foo/bar/baz/": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"name1": "/bar/baz/"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"name1": "/bar/baz"}},
			},
		},
	}})
}

func TestSubtreeMatch(t *testing.T) {
	testPathMatch(t, []pathSpec{{
		/* considering as unspecified */
		`route1: PathSubtree("") -> <shunt>;`,
		map[string]pathSpecTest{
			"": {
				considerTrailing: pathSpecVariant{"route1", nil},
				ignoreTrailing:   pathSpecVariant{"route1", nil},
			},
			"/": {
				considerTrailing: pathSpecVariant{"route1", nil},
				ignoreTrailing:   pathSpecVariant{"route1", nil},
			},
			"/foo": {
				considerTrailing: pathSpecVariant{"route1", nil},
				ignoreTrailing:   pathSpecVariant{"route1", nil},
			},
		},
	}, {
		`route1: PathSubtree("/") -> <shunt>;`,
		map[string]pathSpecTest{
			/* considering as unspecified */
			"": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"*": "/"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"*": "/"}},
			},

			"/": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"*": "/"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"*": "/"}},
			},
			"/foo": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"*": "/foo"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"*": "/foo"}},
			},
		},
	}, {
		`route1: PathSubtree("/foo") -> <shunt>;`,
		map[string]pathSpecTest{
			/* considering as unspecified */
			"": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},

			"/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"*": "/"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"*": "/"}},
			},
			"/foo/": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"*": "/"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"*": "/"}},
			},
			"/foo/bar": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"*": "/bar"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"*": "/bar"}},
			},
			"/foo/bar/": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"*": "/bar/"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"*": "/bar"}},
			},
		},
	}, {
		`route1: PathSubtree("/foo/") -> <shunt>;`,
		map[string]pathSpecTest{
			/* considering as unspecified */
			"": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},

			"/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"*": "/"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"*": "/"}},
			},
			"/foo/": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"*": "/"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"*": "/"}},
			},
			"/foo/bar": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"*": "/bar"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"*": "/bar"}},
			},
			"/foo/bar/": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"*": "/bar/"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"*": "/bar"}},
			},
		},
	}, {
		`route1: PathSubtree("/foo/bar") -> <shunt>;`,
		map[string]pathSpecTest{
			/* considering as unspecified */
			"": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},

			"/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"*": "/"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"*": "/"}},
			},
			"/foo/bar/": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"*": "/"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"*": "/"}},
			},
			"/foo/bar/baz": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"*": "/baz"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"*": "/baz"}},
			},
			"/foo/bar/baz/": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"*": "/baz/"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"*": "/baz"}},
			},
		},
	}, {
		`route1: PathSubtree("/foo/bar/") -> <shunt>;`,
		map[string]pathSpecTest{
			/* considering as unspecified */
			"": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},

			"/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/": {
				considerTrailing: pathSpecVariant{"", nil},
				ignoreTrailing:   pathSpecVariant{"", nil},
			},
			"/foo/bar": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"*": "/"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"*": "/"}},
			},
			"/foo/bar/": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"*": "/"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"*": "/"}},
			},
			"/foo/bar/baz": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"*": "/baz"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"*": "/baz"}},
			},
			"/foo/bar/baz/": {
				considerTrailing: pathSpecVariant{"route1", map[string]string{"*": "/baz/"}},
				ignoreTrailing:   pathSpecVariant{"route1", map[string]string{"*": "/baz"}},
			},
		},
	}})
}

func TestSubtreeWildcardMatch(t *testing.T) {
	/*
	   invalid:

	   PathSubtree(":name1")
	   "", "/", "foo", "/foo", "/foo/", "/foo/bar", "/foo/bar/"

	   PathSubtree(":name1/")
	   "", "/", "foo", "/foo", "/foo/", "/foo/bar", "/foo/bar/"

	   PathSubtree("/foo:name1")
	   "", "/", "foo", "/foo", "/foo/", "/foo/bar", "/foo/bar/"

	   PathSubtree("/foo:name1/")
	   "", "/", "foo", "/foo", "/foo/", "/foo/bar", "/foo/bar/"

	   --

	   PathSubtree("/:name1")
	   "", "/", "/foo", "/foo/", "/foo/bar", "/foo/bar/"

	   PathSubtree("/:name1/")
	   "", "/", "/foo", "/foo/", "/foo/bar", "/foo/bar/"

	   PathSubtree("/:name1/bar")
	   "", "/", "/foo", "/foo/", "/foo/bar", "/foo/bar/", "/foo/bar/baz", "/foo/bar/baz/"

	   PathSubtree("/:name1/bar/")
	   "", "/", "/foo", "/foo/", "/foo/bar", "/foo/bar/", "/foo/bar/baz", "/foo/bar/baz/"

	   PathSubtree("/foo/:name1")
	   "", "/", "/foo", "/foo/", "/foo/bar", "/foo/bar/", "/foo/bar/baz", "/foo/bar/baz/"

	   PathSubtree("/foo/:name1/")
	   "", "/", "/foo", "/foo/", "/foo/bar", "/foo/bar/", "/foo/bar/baz", "/foo/bar/baz/"

	   PathSubtree("/foo/:name1/baz")
	   "", "/", "/foo", "/foo/", "/foo/bar", "/foo/bar/", "/foo/bar/baz", "/foo/bar/baz/", "/foo/bar/baz/qux", "/foo/bar/baz/qux/"

	   PathSubtree("/foo/:name1/baz/")
	   "", "/", "/foo", "/foo/", "/foo/bar", "/foo/bar/", "/foo/bar/baz", "/foo/bar/baz/", "/foo/bar/baz/qux", "/foo/bar/baz/qux/"

	   PathSubtree("/:name1/:name2")
	   "", "/", "/foo", "/foo/", "/foo/bar", "/foo/bar/", "/foo/bar/baz", "/foo/bar/baz/"

	   PathSubtree("/:name1/:name2/")
	   "", "/", "/foo", "/foo/", "/foo/bar", "/foo/bar/", "/foo/bar/baz", "/foo/bar/baz/"

	   PathSubtree("/foo/:name1/:name2")
	   "", "/", "/foo", "/foo/", "/foo/bar", "/foo/bar/", "/foo/bar/baz", "/foo/bar/baz/", "/foo/bar/baz/qux", "/foo/bar/baz/qux/"

	   PathSubtree("/foo/:name1/:name2/")
	   "", "/", "/foo", "/foo/", "/foo/bar", "/foo/bar/", "/foo/bar/baz", "/foo/bar/baz/", "/foo/bar/baz/qux", "/foo/bar/baz/qux/"

	   PathSubtree("/:name1/bar/:name2")
	   "", "/", "/foo", "/foo/", "/foo/bar", "/foo/bar/", "/foo/bar/baz", "/foo/bar/baz/", "/foo/bar/baz/qux", "/foo/bar/baz/qux/"

	   PathSubtree("/:name1/bar/:name2/")
	   "", "/", "/foo", "/foo/", "/foo/bar", "/foo/bar/", "/foo/bar/baz", "/foo/bar/baz/", "/foo/bar/baz/qux", "/foo/bar/baz/qux/"

	   PathSubtree("/:name1/:name2/baz")
	   "", "/", "/foo", "/foo/", "/foo/bar", "/foo/bar/", "/foo/bar/baz", "/foo/bar/baz/", "/foo/bar/baz/qux", "/foo/bar/baz/qux/"

	   PathSubtree("/:name1/:name2/baz/")
	   "", "/", "/foo", "/foo/", "/foo/bar", "/foo/bar/", "/foo/bar/baz", "/foo/bar/baz/", "/foo/bar/baz/qux", "/foo/bar/baz/qux/"
	*/
}

func TestSubtreeCatchallWildcardMatch(t *testing.T) {
	/*
	   invalid:

	   PathSubtree("*")
	   "", "/", "/foo", "/foo/", "/foo/bar", "/foo/bar/"

	   PathSubtree("*name1")
	   "", "/", "/foo", "/foo/", "/foo/bar", "/foo/bar/"

	   PathSubtree("/*")
	   "", "/", "/foo", "/foo/", "/foo/bar", "/foo/bar/"

	   PathSubtree("/foo*")
	   "", "/", "/foo", "/foo/", "/foo/bar", "/foo/bar/"

	   PathSubtree("/foo/*")
	   "", "/", "/foo", "/foo/", "/foo/bar", "/foo/bar/"

	   PathSubtree("/*name1/")
	   "", "/", "/foo", "/foo/", "/foo/bar", "/foo/bar/"

	   PathSubtree("/*name1/bar")
	   "", "/", "/foo", "/foo/", "/foo/bar", "/foo/bar/"

	   PathSubtree("/*name1/bar/")
	   "", "/", "/foo", "/foo/", "/foo/bar", "/foo/bar/"

	   PathSubtree("/foo/*name1/")
	   "", "/", "/foo", "/foo/", "/foo/bar", "/foo/bar/", "/foo/bar/baz", "/foo/bar/baz/"

	   PathSubtree("/foo/*name1/baz")
	   "", "/", "/foo", "/foo/", "/foo/bar", "/foo/bar/", "/foo/bar/baz", "/foo/bar/baz/"

	   --

	   PathSubtree("/*name1")
	   "", "/", "/foo", "/foo/", "/foo/bar", "/foo/bar/"

	   PathSubtree("/foo/*name1")
	   "", "/", "/foo", "/foo/", "/foo/bar", "/foo/bar/", "/foo/bar/baz", "/foo/bar/baz/"
	*/
}
