package restxml

import "net/http"

type Handler struct{}

func (f *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {}

// New returns a new Handler that will dispatch incoming requests to
// one or more fake service backends given as arguments.
func New(serviceBackend interface{}) *Handler {
	return &Handler{}
}
