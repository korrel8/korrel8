package korrel8

import (
	"context"
	"fmt"
	"strings"
)

// Mock implementations

const mockDomain Domain = "mock"

type mockClass string

func (c mockClass) Domain() Domain { return mockDomain }

type mockObject struct {
	name  string
	class mockClass
}

func o(name, class string) Object { return mockObject{name: name, class: mockClass(class)} }
func (o mockObject) Class() Class { return o.class }
func (o mockObject) Native() any  { return o }

// Identifier suppose name is only unique per class (like a K8s name) so identifier is whole object.
func (o mockObject) Identifier() Identifier { return o }

type mockRule struct {
	start, goal Class
	apply       func(Object, *Constraint) Queries
}

func (r mockRule) Start() Class   { return r.start }
func (r mockRule) Goal() Class    { return r.goal }
func (r mockRule) String() string { return fmt.Sprintf("(%v)->%v", r.start, r.goal) }
func (r mockRule) Apply(start Object, c *Constraint) (Queries, error) {
	return r.apply(start, c), nil
}

func rr(start, goal string, apply func(Object, *Constraint) Queries) mockRule {
	return mockRule{
		start: mockClass(start),
		goal:  mockClass(goal),
		apply: apply,
	}
}

func r(start, goal string) mockRule { return rr(start, goal, nil) }

type mockStore struct{}

// Query a mock "query" is a comma-separated list of "name.class" to be turned into mock objects.
func (s mockStore) Query(_ context.Context, q string) ([]Object, error) {
	var objs []Object
	for _, s := range strings.Split(q, ",") {
		nc := strings.Split(s, ".")
		objs = append(objs, mockObject{name: nc[0], class: mockClass(nc[1])})
	}
	return objs, nil
}

var (
	mockStores   = map[Domain]Store{mockDomain: mockStore{}}
	mockFollower = Follower{mockStores}
)