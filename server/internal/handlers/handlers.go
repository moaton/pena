package handlers

type Handlers interface {
	Task()
	Report()
}

type handler struct{}

func NewHandler() Handlers {
	return &handler{}
}

func (h *handler) Task() {

}

func (h *handler) Report() {

}
