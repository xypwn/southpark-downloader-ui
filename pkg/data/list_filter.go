package data

type ListFilter[T any, Pattern any] struct {
	pattern          *Binding[Pattern]
	patternClient    *Client[Pattern]
	parent           *ListBinding[T]
	unfilteredClient *ListClient[T]
	filtered         *ListBinding[T]
	filteredClient   *ListClient[T]
	filterFn         func(value T, pattern Pattern) bool
}

func NewListFilter[T any, Pattern any](parent *ListBinding[T], filterFn func(value T, pattern Pattern) bool) *ListFilter[T, Pattern] {
	res := &ListFilter[T, Pattern]{
		pattern:  NewBinding[Pattern](),
		parent:   parent,
		filterFn: filterFn,
		filtered: NewListBinding[T](),
	}
	res.patternClient = res.pattern.NewClient()
	res.unfilteredClient = parent.NewClient()
	res.filteredClient = res.filtered.NewClient()

	res.filteredClient.AddListener(func([]T) {
		panic("ListFilter: attempted to change filtered data")
	})

	res.pattern.NewClient().AddListener(func(p Pattern) {
		res.updateFilteredData(&p, nil)
	})
	res.parent.NewClient().AddListener(func(data []T) {
		res.updateFilteredData(nil, data)
	})
	return res
}

func (lf *ListFilter[T, Pattern]) Pattern() *Binding[Pattern] {
	return lf.pattern
}

// Read-only! Calling Change() will result in panic.
func (lf *ListFilter[T, Pattern]) Filtered() *ListBinding[T] {
	return lf.filtered
}

func (lf *ListFilter[T, Pattern]) updateFilteredData(newPattern *Pattern, newUnfiltered []T) {
	lf.filteredClient.Change(func(filtered []T) []T {
		filtered = filtered[:0]
		filter := func(pat Pattern, data []T) {
			for _, v := range data {
				if lf.filterFn(v, pat) {
					filtered = append(filtered, v)
				}
			}
		}
		if newPattern != nil {
			lf.unfilteredClient.Examine(func(data []T) {
				filter(*newPattern, data)
			})
		} else if newUnfiltered != nil {
			lf.patternClient.Examine(func(p Pattern) {
				filter(p, newUnfiltered)
			})
		} else {
			panic("updateFilteredData expects either a new pattern or new unfiltered data")
		}
		return filtered
	})
}
